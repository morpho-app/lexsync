package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Commit struct {
	Sha string `json:"sha"`
}

const (
	owner         = "bluesky-social"
	repoName      = "atproto"
	repoURL       = "https://github.com/bluesky-social/atproto.git"
	repoDir       = "/tree/main/lexicons"
	targetRepoURL = "https://github.com/morpho-app/Morpho.git"
	targetRepoDir = "tbd"
)

func main() {
	var lastCommitSha string

	for {
		resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/commits", owner, repoName))
		if err != nil {
			fmt.Println(err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err)
			return
		}

		var commits []Commit
		err = json.Unmarshal(body, &commits)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, commit := range commits {
			if commit.Sha == lastCommitSha {
				break
			}

			// for _, file := range commit.Files {
			// 	if file.Filename == filePath {
			// 		fmt.Println("File has been changed")
			// 		pullRepo()
			// 		// TO-DO
			// 		// convertLexicons()
			// 		pushRepo()
			// 	}
			// }
		}

		if len(commits) > 0 {
			lastCommitSha = commits[0].Sha
		}

		// Wait for a while before checking again
		time.Sleep(24 * time.Hour)
	}
}

func getLatestCommitSha() (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?path=%s", owner, repoName, "lexicons"))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", err
	}

	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found for lexicons directory")
	}

	return commits[0].Sha, nil
}

func pullRepo() error {
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		// Clone the repository if it doesn't exist
		fmt.Println("Cloning repository...")
		cmd := exec.Command("git", "clone", repoURL, repoDir)
		return cmd.Run()
	}

	// Pull the latest changes
	fmt.Println("Pulling latest changes...")
	cmd := exec.Command("git", "-C", repoDir, "pull")
	return cmd.Run()
}

func pushRepo() error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is not set")
	}

	authRepoURL := fmt.Sprintf("https://%s@github.com/morpho-app/Morpho.git", token)

	if _, err := os.Stat(targetRepoDir); os.IsNotExist(err) {
		// Clone the target repository if it doesn't exist
		fmt.Println("Cloning target repository...")
		cmd := exec.Command("git", "clone", authRepoURL, targetRepoDir)
		if err := cmd.Run(); err != nil {
			return err
		}
	} else {
		// Pull the latest changes
		fmt.Println("Pulling latest changes in target repository...")
		cmd := exec.Command("git", "-C", targetRepoDir, "pull")
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Copy the generated Kotlin files to the target repository
	sourceDir := "/lexsync/generated-kotlin"
	// TODO: Double check target repo dir
	destDir := filepath.Join(targetRepoDir, "")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	if err := copyDir(sourceDir, destDir); err != nil {
		return err
	}

	// Add, commit, and push changes
	if err := gitCommitAndPush(targetRepoDir, authRepoURL); err != nil {
		return err
	}

	return nil
}

func gitCommitAndPush(dir, repoURL string) error {
	cmd := exec.Command("git", "-C", dir, "add", ".")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "-C", dir, "commit", "-m", "Syncing new lexicon files")
	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "nothing to commit") {
			fmt.Println("No changes to commit.")
			return nil
		}
		return err
	}

	cmd = exec.Command("git", "-C", dir, "push", repoURL)
	return cmd.Run()
}

// Helper functions
func copyDir(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	return os.WriteFile(dest, input, 0644)
}

// JSON -> Kotlin conversion logic

func mapJSONTypeToKotlin(jsonType string) string {
	switch jsonType {
	case "string":
		return "String"
	case "number":
		return "Double"
	case "integer":
		return "Int"
	case "boolean":
		return "Boolean"
	case "array":
		return "List<Any>"
	case "object":
		return "Map<String, Any>"
	default:
		return "Any"
	}
}

func convertLexicons() error {
	lexiconsDir := filepath.Join(repoDir, "lexicons")
	outputDir := "/lexsync/generated-kotlin"

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	files, err := os.ReadDir(lexiconsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFilePath := filepath.Join(lexiconsDir, file.Name())
			if err := convertJSONToKotlin(jsonFilePath, outputDir); err != nil {
				log.Printf("Error converting %s: %v", file.Name(), err)
			} else {
				fmt.Printf("Converted %s\n", file.Name())
			}
		}
	}

	return nil
}

func convertJSONToKotlin(jsonFilePath, outputDir string) error {
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return err
	}

	className := cases.Title(language.Und).String(strings.TrimSuffix(filepath.Base(jsonFilePath), ".json"))
	kotlinCode := generateKotlinClass(className, data)

	outputFilePath := filepath.Join(outputDir, className+".kt")
	return os.WriteFile(outputFilePath, []byte(kotlinCode), 0644)
}

func generateKotlinClass(className string, data map[string]interface{}) string {
	var fields []string

	if properties, ok := data["properties"].(map[string]interface{}); ok {
		for propName, propValue := range properties {
			propType := "Any"
			if propMap, ok := propValue.(map[string]interface{}); ok {
				if typeName, ok := propMap["type"].(string); ok {
					propType = mapJSONTypeToKotlin(typeName)
				}
			}
			fields = append(fields, fmt.Sprintf("    val %s: %s", propName, propType))
		}
	}

	return fmt.Sprintf("data class %s(\n%s\n)", className, strings.Join(fields, ",\n"))
}

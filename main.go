package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
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

			for _, file := range commit.Files {
				if file.Filename == filePath {
					fmt.Println("File has been changed")
					pullRepo()
					// TO-DO
					// convertLexicons()
					pushRepo()
				}
			}
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

func pushRepo() {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}

	os.Setenv("GIT_ASKPASS", "echo "+token)

	// TO-DO: Add the new files to the repo
	// repo/kotlin-lexicons/* -> https://github.com/morpho-app/Morpho/tree/main/app/src/main/java/app/morpho/lexicons

	cmd := exec.Command("git", "commit", "-m", "Syncing new lexicon files")
	cmd.Dir = ""
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	cmd = exec.Command("git", "push")
	cmd.Dir = ""
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func convertLexicons() {
	// TO-DO: Convert the downloaded lexicon json files to Kotlin data classes
}

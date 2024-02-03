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
	Sha    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
	} `json:"commit"`
	Files []struct {
		Filename string `json:"filename"`
	} `json:"files"`
}

const owner = "bluesky-social"
const repoName = "atproto"
const filePath = "atproto/tree/main/lexicons"
const repoLocation = "https://github.com/bluesky-social/atproto/tree/main/lexicons"

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

func pullRepo() {
	repoURL := repoLocation
	destPath := "/lexsync/repo"

	cmd := exec.Command("git", "clone", repoURL, destPath)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
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

package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// GitHubClient encapsulates methods to interact with the GitHub API.
type GitHubClient struct {
	HTTPClient *http.Client
	Token      string
	UserAgent  string
	Owner      string
	Repo       string
	BaseURL    string
}

func NewGitHubClient(token, userAgent, owner, repo string, baseURL string) *GitHubClient {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	return &GitHubClient{
		HTTPClient: &http.Client{},
		Token:      token,
		UserAgent:  userAgent,
		Owner:      owner,
		Repo:       repo,
		BaseURL:    baseURL,
	}
}

func (c *GitHubClient) GetLatestCommitSha(path string, logger *zap.SugaredLogger) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s/commits?path=%s&per_page=1", c.BaseURL, c.Owner, c.Repo, path)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API responded with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var commits []struct {
		SHA string `json:"sha"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(commits) == 0 {
		return "", fmt.Errorf("no commits found for path: %s", path)
	}

	logger.Infof("Latest commit SHA for path '%s': %s", path, commits[0].SHA)
	return commits[0].SHA, nil
}

func (c *GitHubClient) CloneRepo(repoURL, repoDir string, logger *zap.SugaredLogger) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed or not found in PATH: %v", err)
	}

	if _, err := execCommand("git", "clone", repoURL, repoDir).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	logger.Infof("Repository cloned to %s successfully.", repoDir)
	return nil
}

func (c *GitHubClient) PullLatest(repoDir string, logger *zap.SugaredLogger) error {
	cmd := execCommand("git", "-C", repoDir, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull latest changes: %v, output: %s", err, string(output))
	}

	logger.Infof("Repository at %s updated successfully.", repoDir)
	return nil
}

// CommitAndPush stages all changes, commits them with a predefined message, and pushes to the remote repository.
func (c *GitHubClient) CommitAndPush(repoDir, commitMessage string, logger *zap.SugaredLogger) error {
	// Stage all changes
	cmdAdd := execCommand("git", "-C", repoDir, "add", ".")
	addOutput, err := cmdAdd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage changes: %v, output: %s", err, string(addOutput))
	}

	cmdCommit := execCommand("git", "-C", repoDir, "commit", "-m", commitMessage)
	commitOutput, err := cmdCommit.CombinedOutput()
	if err != nil {
		// If there are no changes to commit, git will exit with a non-zero status
		if strings.Contains(string(commitOutput), "nothing to commit") {
			logger.Info("No changes to commit.")
			return nil
		}
		return fmt.Errorf("failed to commit changes: %v, output: %s", err, string(commitOutput))
	}

	logger.Infof("Changes committed with message: '%s'", commitMessage)

	cmdPush := execCommand("git", "-C", repoDir, "push")
	pushOutput, err := cmdPush.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push changes: %v, output: %s", err, string(pushOutput))
	}

	logger.Info("Changes pushed to the remote repository successfully.")
	return nil
}

// exec.Command wrapper
var execCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

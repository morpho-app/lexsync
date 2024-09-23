package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestGetLatestCommitSha tests the GetLatestCommitSha method of GitHubClient.
func TestGetLatestCommitSha(t *testing.T) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	// Create a mock server to simulate GitHub API responses
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate the request path and query parameters
		expectedPath := "/repos/test_owner/test_repo/commits"
		if r.URL.Path != expectedPath {
			http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
			return
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("path") != "path/to/file" || query.Get("per_page") != "1" {
			http.Error(w, "invalid query parameters", http.StatusBadRequest)
			return
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "token testtoken" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Respond with a mock commit SHA
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"sha":"abcdef1234567890"}]`))
	}))
	defer mockServer.Close()

	// Initialize GitHubClient with the mock server's URL as BaseURL
	client := NewGitHubClient("testtoken", "TestAgent/1.0", "test_owner", "test_repo", mockServer.URL)

	sha, err := client.GetLatestCommitSha("path/to/file", logger)

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, "abcdef1234567890", sha, "Expected SHA to match")
}

// TestGetLatestCommitSha_NoCommits tests GetLatestCommitSha when no commits are found.
func TestGetLatestCommitSha_NoCommits(t *testing.T) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	// Create a mock server that returns an empty commits list
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with an empty array
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	// Initialize GitHubClient with the mock server's URL as BaseURL
	client := NewGitHubClient("testtoken", "TestAgent/1.0", "test_owner", "test_repo", mockServer.URL)

	// Call GetLatestCommitSha
	sha, err := client.GetLatestCommitSha("path/to/file", logger)

	assert.Error(t, err, "Expected an error due to no commits found")
	assert.Equal(t, "", sha, "Expected SHA to be empty")
}

// TestGetLatestCommitSha_InvalidToken tests GetLatestCommitSha with an invalid token.
func TestGetLatestCommitSha_InvalidToken(t *testing.T) {
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	// Create a mock server that simulates unauthorized access
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
	}))
	defer mockServer.Close()

	// Initialize GitHubClient with an invalid token
	client := NewGitHubClient("invalidtoken", "TestAgent/1.0", "test_owner", "test_repo", mockServer.URL)

	// Call GetLatestCommitSha
	sha, err := client.GetLatestCommitSha("path/to/file", logger)

	assert.Error(t, err, "Expected an error due to invalid token")
	assert.Equal(t, "", sha, "Expected SHA to be empty")
}

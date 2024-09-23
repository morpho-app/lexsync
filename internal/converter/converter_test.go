package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

// TestProcessLexiconJSON tests the ProcessLexiconJSON function of the converter.
func TestProcessLexiconJSON(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	logger, _ := config.Build()
	defer logger.Sync()
	sugar := logger.Sugar()

	// Instantiate test cases
	tests := []struct {
		name           string
		jsonContent    string
		expectedOutput string
		expectError    bool
	}{
		{
			name: "Valid_JSON_with_single_definition",
			jsonContent: `{
				"defs": {
					"user": {
						"type": "object",
						"properties": {
							"did": {"type": "string"},
							"handle": {"type": "string"},
							"displayName": {"type": "string"}
						}
					}
				}
			}`,
			expectedOutput: `data class User(
    val did: String?,
    val displayName: String?,
    val handle: String?
)
`,
			expectError: false,
		},
		{
			name: "Valid_JSON_with_multiple_definitions",
			jsonContent: `{
				"defs": {
					"user": {
						"type": "object",
						"properties": {
							"did": {"type": "string"},
							"handle": {"type": "string"},
							"displayName": {"type": "string"}
						}
					},
					"post": {
						"type": "object",
						"properties": {
							"uri": {"type": "string"},
							"cid": {"type": "string"},
							"author": {"$ref": "#/defs/user"},
							"record": {"type": "object"}
						}
					},
					"feed": {
						"type": "object",
						"properties": {
							"posts": {
								"type": "array",
								"items": {"$ref": "#/defs/post"}
							}
						}
					}
				}
			}`,
			expectedOutput: `data class Feed(
    val posts: List<Post>?
)

data class Post(
    val author: User?,
    val cid: String?,
    val record: Map<String, Any>?,
    val uri: String?
)

data class User(
    val did: String?,
    val displayName: String?,
    val handle: String?
)
`,
			expectError: false,
		},
		{
			name: "Invalid_JSON_structure_(missing_'defs')",
			jsonContent: `{
				"components": {}
			}`,
			expectedOutput: "",
			expectError:    true,
		},
		{
			name:           "Malformed_JSON",
			jsonContent:    `{`,
			expectedOutput: "",
			expectError:    true,
		},
		{
			name:           "Empty_JSON",
			jsonContent:    `{}`,
			expectedOutput: "",
			expectError:    true,
		},
		{
			name: "Empty_'defs'",
			jsonContent: `{
				"defs": {}
			}`,
			expectedOutput: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Write the jsonContent to a temporary file
			tempDir := t.TempDir()
			jsonFilePath := filepath.Join(tempDir, "test.json")
			err := os.WriteFile(jsonFilePath, []byte(tc.jsonContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write JSON file: %v", err)
			}

			output, err := ProcessLexiconJSON(jsonFilePath, sugar)
			if tc.expectError {
				assert.Error(t, err, "Expected an error but got none")
				assert.Equal(t, "", output, "Expected empty output on error")
			} else {
				assert.NoError(t, err, "Did not expect an error but got one")
				// Trim both strings to ignore trailing newlines and whitespaces
				trimmedExpected := strings.TrimSpace(tc.expectedOutput)
				trimmedActual := strings.TrimSpace(output)
				assert.Equal(t, trimmedExpected, trimmedActual, "Output mismatch")
			}
		})
	}
}

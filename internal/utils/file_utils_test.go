package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test_ensure_dir")
	defer os.RemoveAll(tmpDir)

	// Directory should not exist initially
	_, err := os.Stat(tmpDir)
	assert.True(t, os.IsNotExist(err), "Directory should not exist before EnsureDir")

	// Ensure the directory
	err = EnsureDir(tmpDir)
	assert.NoError(t, err, "EnsureDir should not return an error")

	// Directory should now exist
	info, err := os.Stat(tmpDir)
	assert.NoError(t, err, "Directory should exist after EnsureDir")
	assert.True(t, info.IsDir(), "EnsureDir should create a directory")
}

func TestCopyFile(t *testing.T) {
	// Create a temporary source file
	srcFile, err := os.CreateTemp("", "src_*.txt")
	assert.NoError(t, err, "Failed to create temporary source file")
	defer os.Remove(srcFile.Name())

	srcContent := []byte("Sample content for testing.")
	_, err = srcFile.Write(srcContent)
	assert.NoError(t, err, "Failed to write to source file")
	srcFile.Close()

	// Create a temporary destination file path
	dstFile := filepath.Join(os.TempDir(), "dst_test_copy.txt")
	defer os.Remove(dstFile)

	// Perform the copy
	err = CopyFile(srcFile.Name(), dstFile)
	assert.NoError(t, err, "CopyFile should not return an error")

	// Read the destination file
	dstContent, err := os.ReadFile(dstFile)
	assert.NoError(t, err, "Failed to read destination file")
	assert.Equal(t, srcContent, dstContent, "Destination file content should match source")
}

func TestCopyDir(t *testing.T) {
	// Create a temporary source directory with nested files
	srcDir, err := os.MkdirTemp("", "src_dir_*")
	assert.NoError(t, err, "Failed to create temporary source directory")
	defer os.RemoveAll(srcDir)

	// Create subdirectories and files
	subDir := filepath.Join(srcDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err, "Failed to create subdirectory")

	file1Path := filepath.Join(srcDir, "file1.txt")
	err = os.WriteFile(file1Path, []byte("Content of file1"), 0644)
	assert.NoError(t, err, "Failed to write file1")

	file2Path := filepath.Join(subDir, "file2.txt")
	err = os.WriteFile(file2Path, []byte("Content of file2"), 0644)
	assert.NoError(t, err, "Failed to write file2")

	// Create a temporary destination directory
	dstDir, err := os.MkdirTemp("", "dst_dir_*")
	assert.NoError(t, err, "Failed to create temporary destination directory")
	defer os.RemoveAll(dstDir)

	// Perform the directory copy
	err = CopyDir(srcDir, dstDir)
	assert.NoError(t, err, "CopyDir should not return an error")

	// Verify that files exist in the destination
	dstFile1Path := filepath.Join(dstDir, "file1.txt")
	dstFile2Path := filepath.Join(dstDir, "subdir", "file2.txt")

	// Check file1
	dstContent1, err := os.ReadFile(dstFile1Path)
	assert.NoError(t, err, "Failed to read destination file1")
	assert.Equal(t, "Content of file1", string(dstContent1), "Destination file1 content should match source")

	// Check file2
	dstContent2, err := os.ReadFile(dstFile2Path)
	assert.NoError(t, err, "Failed to read destination file2")
	assert.Equal(t, "Content of file2", string(dstContent2), "Destination file2 content should match source")
}

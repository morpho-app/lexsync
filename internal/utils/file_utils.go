package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EnsureDir ensures that a directory exists, and creates it if it does not.
func EnsureDir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
	}
	return nil
}

// CopyFile copies a file from src to dst. If dst does not exist, it is created.
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("source file %s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return fmt.Errorf("failed to copy data: %v", err)
	}

	// Optionally, copy file permissions
	if err := os.Chmod(dst, sourceFileStat.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions on destination file: %v", err)
	}

	return nil
}

// ReadFile reads the entire content of the file specified by filePath.
func ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}
	return data, nil
}

// WriteFile writes data to the file specified by filePath.
// If the file does not exist, it is created with the provided permissions.
func WriteFile(filePath string, data []byte, perm os.FileMode) error {
	err := os.WriteFile(filePath, data, perm)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filePath, err)
	}
	return nil
}

// ReadAll reads all data from the provided reader.
func ReadAll(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %v", err)
	}
	return data, nil
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist. Destination directory will be created if it does not exist.
func CopyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source directory: %v", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source %s is not a directory", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %v", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get info for %s: %v", srcPath, err)
		}

		switch mode := entryInfo.Mode(); {
		case mode.IsDir():
			// Recursively copy sub-directories
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		case mode.IsRegular():
			// Copy regular files
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		default:
			// Skip non-regular files (e.g., symlinks, devices)
			fmt.Printf("Skipping non-regular file %s\n", srcPath)
		}
	}

	return nil
}

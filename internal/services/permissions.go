package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// FixSitePermissions fixes permissions for a site directory
func FixSitePermissions(sitePath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(sitePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Fix permissions on parent directories up to home
	if err := fixParentPermissions(absPath); err != nil {
		return fmt.Errorf("failed to fix parent permissions: %w", err)
	}

	// Fix permissions on site directory and contents
	if err := fixDirectoryPermissions(absPath); err != nil {
		return fmt.Errorf("failed to fix directory permissions: %w", err)
	}

	return nil
}

// fixParentPermissions fixes permissions on parent directories
func fixParentPermissions(path string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Walk up to home directory
	current := path
	for {
		// Set directory to 755 (readable/executable by all)
		if err := os.Chmod(current, 0755); err != nil {
			return err
		}

		// Stop at home directory
		if current == homeDir {
			break
		}

		// Move to parent
		parent := filepath.Dir(current)
		if parent == current {
			break // Reached root
		}
		current = parent
	}

	return nil
}

// fixDirectoryPermissions recursively fixes permissions in a directory
func fixDirectoryPermissions(path string) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Directories: 755
			return os.Chmod(filePath, 0755)
		} else {
			// Files: 644
			return os.Chmod(filePath, 0644)
		}
	})
}

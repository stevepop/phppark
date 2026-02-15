package php

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// DetectPHPVersions finds all installed PHP versions
func DetectPHPVersions() ([]PHPVersion, error) {
	return detectLinuxPHP()
}

// detectLinuxPHP finds PHP versions on Linux (Debian/Ubuntu)
func detectLinuxPHP() ([]PHPVersion, error) {
	var versions []PHPVersion
	versionMap := make(map[string]bool) // Deduplicate

	// Common Linux locations
	searchPaths := []string{
		"/usr/bin",
		"/usr/local/bin",
	}

	for _, searchPath := range searchPaths {
		entries, err := os.ReadDir(searchPath)
		if err != nil {
			continue // Directory doesn't exist or can't read
		}

		for _, entry := range entries {
			name := entry.Name()

			// Look for php8.2, php8.3, etc.
			if strings.HasPrefix(name, "php") && len(name) > 3 {
				fullPath := filepath.Join(searchPath, name)

				// Try to get version
				version, err := GetPHPVersionFromBinary(fullPath)
				if err != nil {
					continue
				}

				version = FormatVersion(version)

				// Skip if already found
				if versionMap[version] {
					continue
				}
				versionMap[version] = true

				// Determine FPM socket path
				fpmSocket := fmt.Sprintf("/var/run/php/php%s-fpm.sock", version)

				versions = append(versions, PHPVersion{
					Version:   version,
					FullPath:  fullPath,
					FPMSocket: fpmSocket,
					IsDefault: false,
				})
			}
		}
	}

	// Check for default php
	if defaultPath, err := exec.LookPath("php"); err == nil {
		if version, err := GetPHPVersionFromBinary(defaultPath); err == nil {
			version = FormatVersion(version)

			// Mark the matching version as default
			for i := range versions {
				if versions[i].Version == version {
					versions[i].IsDefault = true
					break
				}
			}
		}
	}

	// Sort by version (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	return versions, nil
}

// ValidatePHPVersion checks if a PHP version is available
func ValidatePHPVersion(version string, availableVersions []PHPVersion) bool {
	for _, v := range availableVersions {
		if v.Version == version {
			return true
		}
	}
	return false
}

// GetDefaultVersion returns the default PHP version
func GetDefaultVersion(versions []PHPVersion) *PHPVersion {
	for i := range versions {
		if versions[i].IsDefault {
			return &versions[i]
		}
	}

	// If no default marked, return first (newest)
	if len(versions) > 0 {
		return &versions[0]
	}

	return nil
}

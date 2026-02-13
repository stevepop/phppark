package php

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// PHPVersion represents an installed PHP version
type PHPVersion struct {
	Version   string // e.g., "8.2"
	FullPath  string // e.g., "/usr/bin/php8.2"
	FPMSocket string // e.g., "/var/run/php/php8.2-fpm.sock"
	IsDefault bool   // Is this the default PHP?
}

// GetPHPVersionFromBinary extracts version from php binary
func GetPHPVersionFromBinary(phpPath string) (string, error) {
	cmd := exec.Command(phpPath, "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse output like "PHP 8.2.15 (cli) ..."
	re := regexp.MustCompile(`PHP (\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse PHP version")
	}

	return matches[1], nil
}

// FormatVersion ensures version is in X.Y format (e.g., "8.2" not "8.2.15")
func FormatVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return version
}

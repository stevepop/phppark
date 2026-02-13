package services

import (
	"fmt"
	"os/exec"
	"strings"
)

// StartPHPFPM starts PHP-FPM service for a given version
func StartPHPFPM(version string) error {
	serviceName := fmt.Sprintf("php%s-fpm", version)

	// Check if running
	cmd := exec.Command("systemctl", "is-active", serviceName)
	if err := cmd.Run(); err == nil {
		return nil // Already running
	}

	// Start service
	cmd = exec.Command("systemctl", "start", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start %s: %w", serviceName, err)
	}

	// Enable on boot
	cmd = exec.Command("systemctl", "enable", serviceName)
	cmd.Run() // Non-fatal

	return nil
}

// EnsurePHPFPMRunning ensures all detected PHP-FPM versions are running
func EnsurePHPFPMRunning(versions []string) error {
	var errors []string

	for _, version := range versions {
		// Extract version number (e.g., "8.2" from "PHP 8.2")
		parts := strings.Split(version, " ")
		if len(parts) >= 2 {
			version = parts[1]
		}

		if err := StartPHPFPM(version); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("PHP-FPM errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// DeployNginxConfig copies config to nginx and reloads
func DeployNginxConfig(siteName, configPath string) error {
	// Paths
	sitesAvailable := "/etc/nginx/sites-available"
	sitesEnabled := "/etc/nginx/sites-enabled"
	defaultSite := filepath.Join(sitesEnabled, "default")

	// Target paths
	availablePath := filepath.Join(sitesAvailable, siteName+".conf")
	enabledPath := filepath.Join(sitesEnabled, siteName+".conf")

	// Copy to sites-available
	if err := copyFile(configPath, availablePath); err != nil {
		return fmt.Errorf("failed to copy config: %w", err)
	}

	// Create symlink in sites-enabled
	if err := createSymlink(availablePath, enabledPath); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Remove default site (first time only)
	if _, err := os.Stat(defaultSite); err == nil {
		if err := os.Remove(defaultSite); err != nil {
			// Non-fatal, just warn
			fmt.Printf("   ⚠️  Could not remove default site: %v\n", err)
		}
	}

	// Test nginx config
	if err := TestNginxConfig(); err != nil {
		return fmt.Errorf("nginx config test failed: %w", err)
	}

	// Reload nginx
	if err := ReloadNginx(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// RemoveNginxConfig removes config from nginx and reloads
func RemoveNginxConfig(siteName string) error {
	sitesAvailable := "/etc/nginx/sites-available"
	sitesEnabled := "/etc/nginx/sites-enabled"

	availablePath := filepath.Join(sitesAvailable, siteName+".conf")
	enabledPath := filepath.Join(sitesEnabled, siteName+".conf")

	// Remove symlink
	if err := os.Remove(enabledPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove enabled config: %w", err)
	}

	// Remove from sites-available
	if err := os.Remove(availablePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove available config: %w", err)
	}

	// Test and reload
	if err := TestNginxConfig(); err != nil {
		return fmt.Errorf("nginx config test failed: %w", err)
	}

	if err := ReloadNginx(); err != nil {
		return fmt.Errorf("failed to reload nginx: %w", err)
	}

	return nil
}

// TestNginxConfig tests nginx configuration
func TestNginxConfig() error {
	cmd := exec.Command("nginx", "-t")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nginx -t failed: %w", err)
	}
	return nil
}

// ReloadNginx reloads nginx service
func ReloadNginx() error {
	cmd := exec.Command("systemctl", "reload", "nginx")
	if err := cmd.Run(); err != nil {
		// Try alternative reload method
		cmd = exec.Command("nginx", "-s", "reload")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to reload nginx: %w", err)
		}
	}
	return nil
}

// StartNginx starts nginx if not running
func StartNginx() error {
	// Check if running
	cmd := exec.Command("systemctl", "is-active", "nginx")
	if err := cmd.Run(); err == nil {
		return nil // Already running
	}

	// Start nginx
	cmd = exec.Command("systemctl", "start", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start nginx: %w", err)
	}

	// Enable on boot
	cmd = exec.Command("systemctl", "enable", "nginx")
	cmd.Run() // Non-fatal

	return nil
}

// Helper: Copy file
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Create directories if needed
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// Helper: Create symlink
func createSymlink(src, dst string) error {
	// Remove existing symlink if it exists
	os.Remove(dst)

	return os.Symlink(src, dst)
}

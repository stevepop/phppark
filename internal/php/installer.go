package php

import (
	"fmt"
	"os/exec"
)

// InstallPHP installs a PHP version with FPM
func InstallPHP(version string) error {
	fmt.Printf("📥 Installing PHP %s-FPM...\n", version)

	// Ensure ondrej PPA is added (for Ubuntu/Debian)
	fmt.Println("   Adding PHP repository...")
	cmd := exec.Command("add-apt-repository", "-y", "ppa:ondrej/php")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add PHP repository: %w", err)
	}

	// Update package list
	fmt.Println("   Updating package list...")
	cmd = exec.Command("apt-get", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update packages: %w", err)
	}

	// Install PHP-FPM
	fmt.Printf("   Installing php%s-fpm...\n", version)
	packageName := fmt.Sprintf("php%s-fpm", version)
	cmd = exec.Command("apt-get", "install", "-y", packageName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install PHP %s: %w", version, err)
	}

	// Install common extensions
	fmt.Println("   Installing common extensions...")
	extensions := []string{
		fmt.Sprintf("php%s-cli", version),
		fmt.Sprintf("php%s-common", version),
		fmt.Sprintf("php%s-mysql", version),
		fmt.Sprintf("php%s-curl", version),
		fmt.Sprintf("php%s-mbstring", version),
		fmt.Sprintf("php%s-xml", version),
		fmt.Sprintf("php%s-zip", version),
	}

	for _, ext := range extensions {
		cmd = exec.Command("apt-get", "install", "-y", ext)
		cmd.Run() // Non-fatal if individual extensions fail
	}

	fmt.Printf("\n✅ PHP %s installed successfully!\n", version)
	return nil
}

// PromptInstallPHP asks user if they want to install a PHP version
func PromptInstallPHP(version string) (bool, error) {
	fmt.Printf("\n⚠️  PHP %s is not installed.\n", version)
	fmt.Printf("   Would you like to install it now? (y/N): ")

	var response string
	fmt.Scanln(&response)

	return response == "y" || response == "Y" || response == "yes", nil
}

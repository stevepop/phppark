package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stevepop/phppark/internal/config"
	"github.com/stevepop/phppark/internal/dns"
	"github.com/stevepop/phppark/internal/nginx"
	"github.com/stevepop/phppark/internal/php"
	"github.com/stevepop/phppark/internal/services"
	"github.com/stevepop/phppark/internal/ssl"
)

var version = "0.1.0-dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "phppark",
		Short:   "PHPark - Development environment manager for Linux",
		Long:    `A modern development environment manager for Linux inspired by Laravel Valet.`,
		Version: version,
	}

	// Add commands
	rootCmd.AddCommand(installCmd())
	rootCmd.AddCommand(setupCmd())
	rootCmd.AddCommand(parkCmd())
	rootCmd.AddCommand(linkCmd())
	rootCmd.AddCommand(unlinkCmd())
	rootCmd.AddCommand(linksCmd())
	rootCmd.AddCommand(rebuildCmd())
	rootCmd.AddCommand(secureCmd())
	rootCmd.AddCommand(unsecureCmd())
	rootCmd.AddCommand(phpListCmd())
	rootCmd.AddCommand(useCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(trustCmd())
	rootCmd.AddCommand(untrustCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func installCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install and configure PHPark",
		Long:  `Install creates the PHPark directory structure and configuration files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall()
		},
	}
}

func runInstall() error {
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	// Check if already installed
	if paths.Exists() {
		fmt.Println("✅ PHPark is already installed!")
		fmt.Printf("\nConfiguration directory: %s\n", paths.Home)
		return nil
	}

	fmt.Println("🚀 Installing PHPark...\n")

	// Create directory structure
	if err := paths.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create default config
	defaultConfig := config.DefaultConfig()
	if err := config.SaveConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create empty sites registry
	emptySites := &config.SiteRegistry{Sites: []config.Site{}}
	if err := config.SaveSites(emptySites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	fmt.Println("✅ PHPark installed successfully!")
	fmt.Printf("\nConfiguration directory: %s\n", paths.Home)
	fmt.Printf("Config file: %s\n", paths.Config)
	fmt.Printf("Sites file: %s\n", paths.Sites)

	fmt.Println("\n🔧 Checking system requirements...")

	missingDeps := []string{}

	// Check for nginx
	if _, err := exec.LookPath("nginx"); err != nil {
		missingDeps = append(missingDeps, "nginx")
	}

	// Check for dnsmasq
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		missingDeps = append(missingDeps, "dnsmasq")
	}

	// Check for PHP
	phpVersions, err := php.DetectPHPVersions()
	if err != nil || len(phpVersions) == 0 {
		missingDeps = append(missingDeps, "PHP-FPM")
	}

	if len(missingDeps) > 0 {
		fmt.Println("\n⚠️  Missing dependencies detected:")
		for _, dep := range missingDeps {
			fmt.Printf("   - %s\n", dep)
		}
		fmt.Println("\n💡 Quick install: Run 'sudo phppark setup' to install everything")
		fmt.Println("   Or install manually: sudo apt install nginx dnsmasq php8.2-fpm")
		return nil
	}

	// Start services if all dependencies present
	fmt.Println("\n🔧 Starting services...")

	if err := services.StartNginx(); err != nil {
		fmt.Printf("⚠️  Warning: Could not start nginx: %v\n", err)
	} else {
		fmt.Println("✅ Nginx started")
	}

	if len(phpVersions) > 0 {
		for _, v := range phpVersions {
			if err := services.StartPHPFPM(v.Version); err == nil {
				fmt.Printf("✅ PHP %s-FPM started\n", v.Version)
			}
		}
	}

	fmt.Println("\n📚 Next steps:")
	fmt.Println("  1. Review/edit config: cat ~/.phppark/config.yaml")
	fmt.Println("  2. Park a directory: phppark park ~/sites")
	fmt.Println("  3. Link a site: phppark link myapp")

	return nil
}

func setupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Complete PHPark setup (install all dependencies)",
		Long:  `Setup installs PHPark and all required dependencies (nginx, dnsmasq, PHP).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}
}

func runSetup() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("setup must be run as root: use 'sudo phppark setup'")
	}

	fmt.Println("🚀 PHPark Complete Setup")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("\nThis will install:")
	fmt.Println("  • nginx (web server)")
	fmt.Println("  • dnsmasq (DNS resolver)")
	fmt.Println("  • PHP 8.2-FPM (with common extensions)")
	fmt.Println("  • PHPark configuration")
	fmt.Printf("\nContinue? (Y/n): ")

	var response string
	fmt.Scanln(&response)

	if response != "" && response != "y" && response != "Y" && response != "yes" {
		fmt.Println("Setup cancelled")
		return nil
	}

	// Update package list first
	fmt.Println("\n📦 Updating package list...")
	cmd := exec.Command("apt-get", "update")
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: apt-get update failed: %v\n", err)
	}

	// Install nginx
	fmt.Println("\n📦 Installing nginx...")
	cmd = exec.Command("apt-get", "install", "-y", "nginx")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install nginx: %w", err)
	}
	fmt.Println("✅ Nginx installed")

	// Install dnsmasq
	fmt.Println("\n📦 Installing dnsmasq...")
	cmd = exec.Command("apt-get", "install", "-y", "dnsmasq")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install dnsmasq: %w", err)
	}
	fmt.Println("✅ dnsmasq installed")

	// Disable systemd-resolved's stub listener if it is occupying port 53.
	// We only disable the stub — systemd-resolved keeps running so that VPN,
	// DHCP, and NetworkManager DNS routing continue to work normally.
	if dns.CheckSystemdResolvedConflict() {
		fmt.Println("\n⚠️  systemd-resolved stub listener is occupying port 53")
		fmt.Println("   Disabling stub listener (systemd-resolved will keep running)...")
		if err := dns.DisableSystemdResolvedStub(); err != nil {
			fmt.Printf("   ⚠️  Warning: could not fix automatically: %v\n", err)
			fmt.Println("   To fix manually, add DNSStubListener=no to /etc/systemd/resolved.conf")
			fmt.Println("   then run: sudo systemctl restart systemd-resolved")
		} else {
			fmt.Println("   ✅ Stub listener disabled — systemd-resolved still running for VPN/DHCP DNS")
		}
	}

	// Install software-properties-common (for add-apt-repository)
	fmt.Println("\n📦 Installing prerequisites...")
	cmd = exec.Command("apt-get", "install", "-y", "software-properties-common")
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠️  Warning: Could not install software-properties-common: %v\n", err)
	}

	// Install PHP 8.2
	fmt.Println("\n📦 Installing PHP 8.2-FPM...")
	if err := php.InstallPHP("8.2"); err != nil {
		return fmt.Errorf("failed to install PHP: %w", err)
	}

	// Initialize PHPark
	fmt.Println("\n🔧 Configuring PHPark...")

	// Create directories
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	if err := paths.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create default config
	defaultConfig := config.DefaultConfig()
	defaultConfig.DefaultPHP = "8.2"
	if err := config.SaveConfig(defaultConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Create empty sites registry
	emptySites := &config.SiteRegistry{Sites: []config.Site{}}
	if err := config.SaveSites(emptySites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	// Start services
	fmt.Println("\n🔧 Starting services...")

	if err := services.StartNginx(); err != nil {
		fmt.Printf("⚠️  Warning: Could not start nginx: %v\n", err)
	} else {
		fmt.Println("✅ Nginx started")
	}

	if err := services.StartPHPFPM("8.2"); err != nil {
		fmt.Printf("⚠️  Warning: Could not start PHP-FPM: %v\n", err)
	} else {
		fmt.Println("✅ PHP 8.2-FPM started")
	}

	// Success message
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Setup complete!")
	fmt.Println(strings.Repeat("=", 50))

	fmt.Printf("\nConfiguration directory: %s\n", paths.Home)
	fmt.Printf("Default PHP version: 8.2\n")

	fmt.Println("\n📚 Try it out:")
	fmt.Println("  mkdir -p ~/sites/myapp/public")
	fmt.Println("  echo '<?php phpinfo(); ?>' > ~/sites/myapp/public/index.php")
	fmt.Println("  cd ~/sites")
	fmt.Println("  sudo phppark park")
	fmt.Println("  curl -H 'Host: myapp.test' http://localhost")

	fmt.Println("\n💡 Tip: Run 'phppark status' to see your configuration")

	return nil
}

func parkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "park [path]",
		Short: "Park a directory - serve all subdirectories as sites",
		Long:  `Park registers a directory so all subdirectories are served as <dirname>.test`,
		Args:  cobra.MaximumNArgs(1), // 0 or 1 argument
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runPark(path)
		},
	}
}

func runPark(path string) error {
	// If no path provided, use current directory
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		fmt.Printf("💡 No path provided, using current directory\n")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Read all subdirectories
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Load existing sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Load config for defaults
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Track what we're adding
	added := 0
	skipped := 0
	var addedSites []string

	fmt.Printf("📦 Parking directory: %s\n\n", absPath)

	// Process each subdirectory
	for _, entry := range entries {
		// Skip non-directories
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories (start with .)
		name := entry.Name()
		if name[0] == '.' {
			continue
		}

		// Check if site already exists
		if existing := sites.FindSite(name); existing != nil {
			fmt.Printf("⏭️  Skipping '%s' (already exists as %s)\n", name, existing.Type)
			skipped++
			continue
		}

		// Create site
		sitePath := filepath.Join(absPath, name)
		site := config.Site{
			Name:       name,
			Path:       sitePath,
			Type:       "park",
			PHPVersion: "", // Use default
			Secured:    cfg.UseHTTPS,
		}

		// Add to registry
		sites.AddSite(site)

		// Generate nginx config
		if err := generateNginxConfig(&site, cfg); err != nil {
			fmt.Printf("⚠️  %s: failed to generate config (%v)\n", name, err)
		} else {
			addedSites = append(addedSites, name)
			added++
		}
	}

	// Save if we added anything
	if added > 0 {
		if err := config.SaveSites(sites); err != nil {
			return fmt.Errorf("failed to save sites: %w", err)
		}
	}

	// Summary
	fmt.Println()
	if added == 0 {
		fmt.Println("⚠️  No new sites added")
		if skipped > 0 {
			fmt.Printf("   %d subdirectories already registered\n", skipped)
		} else {
			fmt.Println("   No subdirectories found in this directory")
		}
	} else {
		fmt.Printf("✅ Parked %d site(s):\n", added)
		for _, name := range addedSites {
			fmt.Printf("   • %s.%s\n", name, cfg.Domain)
		}

		if skipped > 0 {
			fmt.Printf("\n⏭️  Skipped %d existing site(s)\n", skipped)
		}
	}

	return nil
}

func linkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link [name]",
		Short: "Link current directory as a site",
		Long:  `Link creates a site that serves the current directory as <name>.test`,
		Args:  cobra.MaximumNArgs(1), // 0 or 1 argument
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) > 0 {
				name = args[0]
			}
			return runLink(name)
		},
	}
}

func runLink(name string) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// If no name provided, use directory name
	if name == "" {
		name = filepath.Base(currentDir)
		fmt.Printf("💡 No name provided, using directory name: %s\n", name)
	}

	// Load existing sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Check if site already exists
	if existing := sites.FindSite(name); existing != nil {
		fmt.Printf("⚠️  Site '%s' already exists:\n", name)
		fmt.Printf("   Current path: %s\n", existing.Path)
		fmt.Printf("   New path:     %s\n", currentDir)
		fmt.Println("\nTo update, unlink first: phppark unlink", name)
		return nil
	}

	// Load config to get default PHP
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create new site
	site := config.Site{
		Name:       name,
		Path:       currentDir,
		Type:       "link",
		PHPVersion: "", // Use default from config
		Secured:    cfg.UseHTTPS,
	}

	// Add site to registry
	sites.AddSite(site)

	// Save registry
	if err := config.SaveSites(sites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	// Generate nginx config
	fmt.Printf("✅ Linked site: %s.%s\n", name, cfg.Domain)
	fmt.Printf("   Path: %s\n", currentDir)

	if err := generateNginxConfig(&site, cfg); err != nil {
		fmt.Printf("   ⚠️  Warning: %v\n", err)
		fmt.Println("   Site registered but nginx config not created")
	} else {
		fmt.Println("   ✅ Nginx config generated")
	}

	// Rest of success message
	phpVersion := cfg.DefaultPHP
	if site.PHPVersion != "" {
		phpVersion = site.PHPVersion
	}
	fmt.Printf("   PHP:  %s\n", phpVersion)

	return nil
}

func unlinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlink [name]",
		Short: "Remove a linked site",
		Long:  `Unlink removes a site from PHPark management.`,
		Args:  cobra.ExactArgs(1), // Exactly 1 argument required
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlink(args[0])
		},
	}
}

func runUnlink(siteName string) error {
	// Load sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Find site
	site := sites.FindSite(siteName)
	if site == nil {
		return fmt.Errorf("site '%s' not found", siteName)
	}

	// Get config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Display info
	fmt.Printf("🗑️  Removing site: %s.%s\n", siteName, cfg.Domain)
	fmt.Printf("   Path: %s\n", site.Path)
	fmt.Printf("   Type: %s\n", site.Type)

	// Get paths
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	// Remove nginx config file
	configPath := filepath.Join(paths.Nginx, siteName+".conf")
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}
	fmt.Println("   🗑️  Removed nginx config")

	if err := services.RemoveNginxConfig(siteName); err != nil {
		fmt.Printf("   ⚠️  Warning: Could not remove from nginx: %v\n", err)
	} else {
		fmt.Println("   ✅ Removed from nginx")
	}

	// Remove from registry
	sites.RemoveSite(siteName)
	if err := config.SaveSites(sites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	fmt.Println("\n✅ Site unlinked successfully")

	return nil
}

func linksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "links",
		Short: "List all linked sites",
		Long:  `List displays all parked and linked sites managed by PHPark.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinks()
		},
	}
}

func runLinks() error {
	// Load sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Check if empty
	allSites := sites.ListSites()
	if len(allSites) == 0 {
		fmt.Println("📋 No sites registered yet.")
		fmt.Println("\nTo add sites:")
		fmt.Println("  phppark park ~/sites    # Park a directory")
		fmt.Println("  phppark link myapp      # Link current directory")
		return nil
	}

	// Display sites
	fmt.Printf("📋 Registered Sites (%d total)\n\n", len(allSites))

	for _, site := range allSites {
		// Site name and URL
		fmt.Printf("🔗 %s.test\n", site.Name)

		// Path
		fmt.Printf("   Path: %s\n", site.Path)

		// Type
		typeIcon := "📌"
		if site.Type == "park" {
			typeIcon = "📦"
		}
		fmt.Printf("   Type: %s %s\n", typeIcon, site.Type)

		// PHP version
		phpVersion := site.PHPVersion
		if phpVersion == "" {
			phpVersion = "(default)"
		}
		fmt.Printf("   PHP:  %s\n", phpVersion)

		// HTTPS status
		httpsStatus := "❌ HTTP"
		if site.Secured {
			httpsStatus = "✅ HTTPS"
		}
		fmt.Printf("   SSL:  %s\n", httpsStatus)

		fmt.Println() // Empty line between sites
	}

	return nil
}

func generateNginxConfig(site *config.Site, cfg *config.Config) error {
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	// Determine PHP version
	phpVersion := site.PHPVersion
	if phpVersion == "" {
		phpVersion = cfg.DefaultPHP
	}

	// Create site config
	nginxCfg := nginx.CreateSiteConfig(
		site.Name,    // siteName
		site.Path,    // sitePath
		cfg.Domain,   // domain
		phpVersion,   // phpVersion
		site.Secured, // useSSL
	)

	// If secured, add certificate paths
	if site.Secured {
		nginxCfg.CertPath = filepath.Join(paths.Certificates, site.Name+".crt")
		nginxCfg.KeyPath = filepath.Join(paths.Certificates, site.Name+".key")
	}

	// Generate config content
	configContent, err := nginx.GenerateConfig(nginxCfg)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write to file
	configPath := filepath.Join(paths.Nginx, site.Name+".conf")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("   📄 Config: %s\n", configPath)

	// Fix permissions first
	if err := services.FixSitePermissions(site.Path); err != nil {
		fmt.Printf("   ⚠️  Warning: Could not fix permissions: %v\n", err)
	}

	// Deploy to nginx
	if err := services.DeployNginxConfig(site.Name, configPath); err != nil {
		fmt.Printf("   ⚠️  Warning: Could not deploy to nginx: %v\n", err)
		fmt.Println("   Run manually: sudo cp ~/.phppark/nginx/*.conf /etc/nginx/sites-available/")
	} else {
		fmt.Printf("   ✅ Deployed to nginx\n")
	}

	// Start PHP-FPM
	if phpVersion != "" {
		if err := services.StartPHPFPM(phpVersion); err != nil {
			fmt.Printf("   ⚠️  Warning: Could not start PHP-FPM: %v\n", err)
		}
	}

	// Ensure nginx is running
	if err := services.StartNginx(); err != nil {
		fmt.Printf("   ⚠️  Warning: Could not start nginx: %v\n", err)
	}

	return nil
}

func rebuildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild all nginx configurations",
		Long:  `Rebuild regenerates nginx configuration files for all registered sites.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRebuild()
		},
	}
}

func runRebuild() error {
	// Load sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	allSites := sites.ListSites()
	if len(allSites) == 0 {
		fmt.Println("📋 No sites to rebuild")
		return nil
	}

	fmt.Printf("🔨 Rebuilding nginx configs for %d site(s)...\n\n", len(allSites))

	success := 0
	failed := 0

	for _, site := range allSites {
		fmt.Printf("   %s.%s ... ", site.Name, cfg.Domain)

		if err := generateNginxConfig(&site, cfg); err != nil {
			fmt.Printf("❌ failed (%v)\n", err)
			failed++
		} else {
			fmt.Printf("✅\n")
			success++
		}
	}

	fmt.Printf("\n✅ Rebuilt %d config(s)", success)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()

	return nil
}

func secureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "secure [site]",
		Short: "Enable HTTPS for a site",
		Long:  `Secure generates SSL certificates and enables HTTPS for a site.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSecure(args[0])
		},
	}
}

func runSecure(siteName string) error {
	// Load sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Find site
	site := sites.FindSite(siteName)
	if site == nil {
		return fmt.Errorf("site '%s' not found", siteName)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get paths
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	fmt.Printf("🔒 Securing %s.%s...\n", siteName, cfg.Domain)

	// Check if already secured
	if site.Secured {
		fmt.Println("   ⚠️  Site is already secured")

		// Check if certs exist
		if ssl.CertificateExists(siteName, paths.Certificates) {
			fmt.Println("   Certificates already exist")
			return nil
		}

		fmt.Println("   Regenerating certificates...")
	}

	// Generate certificates
	certPaths, err := ssl.GenerateSelfSignedCert(siteName, cfg.Domain, paths.Certificates)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	fmt.Printf("   📜 Certificate: %s\n", certPaths.CertFile)
	fmt.Printf("   🔑 Private Key: %s\n", certPaths.KeyFile)

	// Update site to be secured
	site.Secured = true
	sites.AddSite(*site) // Updates existing

	// Save sites
	if err := config.SaveSites(sites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	// Regenerate nginx config with SSL
	if err := generateNginxConfig(site, cfg); err != nil {
		return fmt.Errorf("failed to update nginx config: %w", err)
	}

	fmt.Println("\n✅ Site secured successfully!")
	fmt.Printf("   Access via: https://%s.%s\n", siteName, cfg.Domain)
	fmt.Println("\n⚠️  Note: You may need to accept the self-signed certificate in your browser")

	return nil
}

func unsecureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unsecure [site]",
		Short: "Disable HTTPS for a site",
		Long:  `Unsecure removes SSL certificates and disables HTTPS for a site.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnsecure(args[0])
		},
	}
}

func runUnsecure(siteName string) error {
	// Load sites
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	// Find site
	site := sites.FindSite(siteName)
	if site == nil {
		return fmt.Errorf("site '%s' not found", siteName)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get paths
	paths, err := config.GetPaths()
	if err != nil {
		return err
	}

	fmt.Printf("🔓 Unsecuring %s.%s...\n", siteName, cfg.Domain)

	// Check if not secured
	if !site.Secured {
		fmt.Println("   ⚠️  Site is not secured (already HTTP)")
		return nil
	}

	// Remove certificates
	if err := ssl.RemoveCertificate(siteName, paths.Certificates); err != nil {
		fmt.Printf("   ⚠️  Warning: failed to remove certificates: %v\n", err)
	} else {
		fmt.Println("   🗑️  Removed SSL certificates")
	}

	// Update site to be unsecured
	site.Secured = false
	sites.AddSite(*site) // Updates existing

	// Save sites
	if err := config.SaveSites(sites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	// Regenerate nginx config without SSL
	if err := generateNginxConfig(site, cfg); err != nil {
		return fmt.Errorf("failed to update nginx config: %w", err)
	}

	fmt.Println("\n✅ Site unsecured successfully!")
	fmt.Printf("   Access via: http://%s.%s\n", siteName, cfg.Domain)

	return nil
}

func phpListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "php:list",
		Short: "List installed PHP versions",
		Long:  `List all PHP versions detected on the system.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPHPList()
		},
	}
}

func runPHPList() error {
	fmt.Println("🔍 Detecting PHP versions...\n")

	versions, err := php.DetectPHPVersions()
	if err != nil {
		return fmt.Errorf("failed to detect PHP versions: %w", err)
	}

	if len(versions) == 0 {
		fmt.Println("❌ No PHP installations found")
		fmt.Println("\nPlease install PHP to use PHPark")
		return nil
	}

	fmt.Printf("Found %d PHP version(s):\n\n", len(versions))

	for _, v := range versions {
		marker := "  "
		if v.IsDefault {
			marker = "✓ "
		}

		fmt.Printf("%s PHP %s\n", marker, v.Version)
		fmt.Printf("   Binary: %s\n", v.FullPath)
		fmt.Printf("   Socket: %s\n", v.FPMSocket)

		if v.IsDefault {
			fmt.Printf("   Status: Default\n")
		}

		fmt.Println()
	}

	return nil
}

func useCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <php-version> [site]",
		Short: "Set PHP version for a site (or globally)",
		Long:  `Use sets the PHP version for a specific site, or updates the default if no site specified.`,
		Args:  cobra.RangeArgs(1, 2), // 1 or 2 arguments
		RunE: func(cmd *cobra.Command, args []string) error {
			phpVersion := args[0]
			siteName := ""
			if len(args) > 1 {
				siteName = args[1]
			}
			return runUse(phpVersion, siteName)
		},
	}
}

func runUse(phpVersion, siteName string) error {
	// Detect available PHP versions
	versions, err := php.DetectPHPVersions()
	if err != nil {
		return fmt.Errorf("failed to detect PHP versions: %w", err)
	}

	// Format version (allow "8.2" or just "8.2")
	phpVersion = php.FormatVersion(phpVersion)

	// Check if version exists
	versionExists := php.ValidatePHPVersion(phpVersion, versions)

	if !versionExists {
		fmt.Printf("❌ PHP %s is not installed\n\n", phpVersion)

		// Show available versions
		if len(versions) > 0 {
			fmt.Println("Available versions:")
			for _, v := range versions {
				fmt.Printf("  - %s\n", v.Version)
			}
			fmt.Println()
		}

		shouldInstall, err := php.PromptInstallPHP(phpVersion)
		if err != nil {
			return err
		}

		if shouldInstall {
			if err := php.InstallPHP(phpVersion); err != nil {
				return fmt.Errorf("installation failed: %w", err)
			}

			// Re-detect versions
			versions, err = php.DetectPHPVersions()
			if err != nil {
				return fmt.Errorf("failed to detect PHP versions: %w", err)
			}

			// Verify installation
			if !php.ValidatePHPVersion(phpVersion, versions) {
				return fmt.Errorf("installation completed but PHP %s not detected", phpVersion)
			}

			fmt.Printf("\n✅ PHP %s is now available!\n\n", phpVersion)
		} else {
			return fmt.Errorf("PHP %s is required but not installed", phpVersion)
		}
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no site specified, update global default
	if siteName == "" {
		cfg.DefaultPHP = phpVersion
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✅ Set default PHP version to %s\n", phpVersion)

		// Switch CLI PHP version
		phpPath := fmt.Sprintf("/usr/bin/php%s", phpVersion)
		cmd := exec.Command("update-alternatives", "--set", "php", phpPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("\n⚠️  Warning: Could not update CLI PHP: %v\n", err)
			fmt.Printf("   Sites will use PHP %s via PHP-FPM\n", phpVersion)
			fmt.Printf("   To manually switch CLI: sudo update-alternatives --set php %s\n", phpPath)
		} else {
			fmt.Printf("   ✅ CLI PHP switched to %s\n", phpVersion)
		}

		fmt.Println("\nNew sites will use PHP", phpVersion)
		fmt.Println("To update existing sites, run: sudo phppark rebuild")
		fmt.Println("\n💡 Verify CLI change: php -v")

		return nil
	}

	// Update specific site
	sites, err := config.LoadSites()
	if err != nil {
		return fmt.Errorf("failed to load sites: %w", err)
	}

	site := sites.FindSite(siteName)
	if site == nil {
		return fmt.Errorf("site '%s' not found", siteName)
	}

	// Update site's PHP version
	site.PHPVersion = phpVersion
	sites.AddSite(*site)

	if err := config.SaveSites(sites); err != nil {
		return fmt.Errorf("failed to save sites: %w", err)
	}

	fmt.Printf("✅ Set PHP %s for %s.%s\n", phpVersion, siteName, cfg.Domain)
	fmt.Println("\n⚠️  Note: Run 'sudo phppark rebuild' to apply changes")

	return nil
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show PHPark installation status",
		Long:  `Status displays the current PHPark configuration and system status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func runStatus() error {
	fmt.Println("📊 PHPark Status\n")

	// Check if PHPark is installed
	paths, err := config.GetPaths()
	if err != nil {
		return fmt.Errorf("failed to get paths: %w", err)
	}

	// Installation Status
	fmt.Println("=== Installation ===")
	if paths.Exists() {
		fmt.Printf("✅ PHPark is installed at %s\n", paths.Home)
	} else {
		fmt.Printf("❌ PHPark is not installed\n")
		fmt.Println("   Run: phppark install")
		return nil
	}

	// Configuration
	fmt.Println("\n=== Configuration ===")
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("⚠️  Failed to load config: %v\n", err)
	} else {
		fmt.Printf("Domain:      .%s\n", cfg.Domain)
		fmt.Printf("Default PHP: %s\n", cfg.DefaultPHP)
		fmt.Printf("HTTPS:       %v\n", cfg.UseHTTPS)
		fmt.Printf("Config:      %s\n", paths.Config)
	}

	// Sites
	fmt.Println("\n=== Sites ===")
	sites, err := config.LoadSites()
	if err != nil {
		fmt.Printf("⚠️  Failed to load sites: %v\n", err)
	} else {
		allSites := sites.ListSites()
		fmt.Printf("Total sites: %d\n", len(allSites))

		// Count by type
		linked := 0
		parked := 0
		secured := 0
		for _, site := range allSites {
			if site.Type == "link" {
				linked++
			} else {
				parked++
			}
			if site.Secured {
				secured++
			}
		}

		fmt.Printf("  Linked:    %d\n", linked)
		fmt.Printf("  Parked:    %d\n", parked)
		fmt.Printf("  Secured:   %d (HTTPS)\n", secured)
		fmt.Printf("Registry:    %s\n", paths.Sites)
	}

	// Nginx Configs
	fmt.Println("\n=== Nginx ===")
	nginxConfigs, err := os.ReadDir(paths.Nginx)
	if err != nil {
		fmt.Printf("⚠️  Failed to read nginx configs: %v\n", err)
	} else {
		configCount := 0
		for _, entry := range nginxConfigs {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".conf" {
				configCount++
			}
		}
		fmt.Printf("Configs:     %d generated\n", configCount)
		fmt.Printf("Location:    %s\n", paths.Nginx)
	}

	// SSL Certificates
	fmt.Println("\n=== SSL Certificates ===")
	certs, err := os.ReadDir(paths.Certificates)
	if err != nil {
		fmt.Printf("⚠️  Failed to read certificates: %v\n", err)
	} else {
		certCount := 0
		for _, entry := range certs {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".crt" {
				certCount++
			}
		}
		fmt.Printf("Certificates: %d\n", certCount)
		fmt.Printf("Location:     %s\n", paths.Certificates)
	}

	// PHP Versions
	fmt.Println("\n=== PHP ===")
	phpVersions, err := php.DetectPHPVersions()
	if err != nil {
		fmt.Printf("⚠️  Failed to detect PHP: %v\n", err)
	} else {
		if len(phpVersions) == 0 {
			fmt.Println("❌ No PHP installations found")
		} else {
			fmt.Printf("Installed:   %d version(s)\n", len(phpVersions))
			for _, v := range phpVersions {
				marker := "  "
				if v.IsDefault {
					marker = "✓ "
				}
				fmt.Printf("%sPHP %s (%s)\n", marker, v.Version, v.FullPath)
			}
		}
	}

	// System Info
	fmt.Println("\n=== System ===")
	fmt.Printf("OS:          %s\n", runtime.GOOS)
	fmt.Printf("Arch:        %s\n", runtime.GOARCH)

	// Check for nginx
	if _, err := exec.LookPath("nginx"); err == nil {
		cmd := exec.Command("nginx", "-v")
		output, _ := cmd.CombinedOutput()
		fmt.Printf("Nginx:       ✅ %s\n", strings.TrimSpace(string(output)))
	} else {
		fmt.Println("Nginx:       ❌ Not found")
	}

	// Check for dnsmasq
	if _, err := exec.LookPath("dnsmasq"); err == nil {
		fmt.Println("dnsmasq:     ✅ Installed")
	} else {
		fmt.Println("dnsmasq:     ❌ Not found")
	}

	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Run 'phppark links' to see all registered sites")

	// DNS Configuration
	fmt.Println("\n=== DNS ===")
	isConfigured, err := dns.CheckDNS(cfg.Domain)
	if err != nil {
		fmt.Printf("⚠️  Failed to check DNS: %v\n", err)
	} else {
		if isConfigured {
			fmt.Printf("Status:      ✅ Configured for .%s\n", cfg.Domain)
		} else {
			fmt.Printf("Status:      ❌ Not configured\n")
			fmt.Println("Setup:       Run 'phppark trust'")
		}
	}

	return nil
}

func trustCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trust",
		Short: "Setup DNS resolution for .test domains",
		Long:  `Trust configures your system to resolve .test domains to localhost.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrust()
		},
	}
}

func runTrust() error {
	// Load config to get domain
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("🔧 Configuring DNS for .%s domains...\n\n", cfg.Domain)

	// Check if already configured
	isConfigured, err := dns.CheckDNS(cfg.Domain)
	if err != nil {
		return fmt.Errorf("failed to check DNS: %w", err)
	}

	// Check for systemd-resolved stub listener conflict regardless of whether
	// the dnsmasq config file already exists. A previous failed run may have
	// written the config without ever freeing port 53.
	if dns.CheckSystemdResolvedConflict() {
		fmt.Println("\n⚠️  systemd-resolved stub listener is occupying port 53")
		fmt.Println("   This is common on Ubuntu/Debian systems (including EC2 instances).")
		fmt.Println("   PHPark can disable the stub listener only — systemd-resolved will keep")
		fmt.Println("   running, so VPN routing, DHCP DNS, and NetworkManager continue to work.")
		fmt.Printf("   Disable stub listener now? (Y/n): ")
		var ans string
		fmt.Scanln(&ans)
		if ans == "" || ans == "y" || ans == "Y" || ans == "yes" {
			if err := dns.DisableSystemdResolvedStub(); err != nil {
				fmt.Printf("   ⚠️  Warning: %v\n", err)
				fmt.Println("   To fix manually, add DNSStubListener=no to /etc/systemd/resolved.conf")
				fmt.Println("   then run: sudo systemctl restart systemd-resolved")
			} else {
				fmt.Println("   ✅ Stub listener disabled — systemd-resolved still running for VPN/DHCP DNS\n")
			}
		}
	}

	if isConfigured {
		fmt.Printf("✅ DNS resolver is configured for .%s\n", cfg.Domain)
	} else {
		fmt.Println("Setting up dnsmasq...")
		fmt.Println("⚠️  This requires sudo access")

		if err := dns.SetupDNS(cfg.Domain); err != nil {
			return fmt.Errorf("failed to setup DNS: %w", err)
		}

		fmt.Printf("\n✅ DNS configured for .%s domains\n", cfg.Domain)
	}

	// Always ensure dnsmasq is running — the config file may exist from a
	// previous partial run where the service never successfully started.
	if err := exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run(); err != nil {
		fmt.Printf("⚠️  Warning: could not restart dnsmasq: %v\n", err)
	} else {
		fmt.Println("✅ dnsmasq running")
	}

	fmt.Println("\nTesting resolution...")

	// Test resolution
	fmt.Println("\n=== Testing DNS Resolution ===")

	// Load a few sites to test
	sites, err := config.LoadSites()
	if err == nil && len(sites.ListSites()) > 0 {
		// Test first 3 sites
		testCount := 3
		if len(sites.ListSites()) < testCount {
			testCount = len(sites.ListSites())
		}

		for i := 0; i < testCount; i++ {
			site := sites.ListSites()[i]
			hostname := fmt.Sprintf("%s.%s", site.Name, cfg.Domain)

			fmt.Printf("Testing %s ... ", hostname)

			resolves, err := dns.TestDNSResolution(hostname)
			if err != nil {
				fmt.Println("❌ Error")
			} else if resolves {
				fmt.Println("✅ Resolves to 127.0.0.1")
			} else {
				fmt.Println("⚠️  Does not resolve (may need to wait for cache)")
			}
		}
	} else {
		// Test with example
		testHost := fmt.Sprintf("example.%s", cfg.Domain)
		fmt.Printf("Testing %s ... ", testHost)

		resolves, err := dns.TestDNSResolution(testHost)
		if err != nil {
			fmt.Println("❌ Error")
		} else if resolves {
			fmt.Println("✅ Resolves to 127.0.0.1")
		} else {
			fmt.Println("⚠️  Does not resolve (may need to wait for cache)")
		}
	}

	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("✅ DNS setup complete!")
	fmt.Printf("All .%s domains now resolve to localhost\n", cfg.Domain)

	return nil
}

func untrustCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "untrust",
		Short: "Remove DNS resolution for .test domains",
		Long:  `Untrust removes DNS configuration for .test domains.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUntrust()
		},
	}
}

func runUntrust() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("🔧 Removing DNS configuration for .%s domains...\n", cfg.Domain)
	fmt.Println("⚠️  This requires sudo access")

	if err := dns.RemoveDNS(cfg.Domain); err != nil {
		return fmt.Errorf("failed to remove DNS: %w", err)
	}

	fmt.Printf("\n✅ DNS configuration removed for .%s\n", cfg.Domain)
	fmt.Println("Sites will no longer resolve automatically")

	return nil
}

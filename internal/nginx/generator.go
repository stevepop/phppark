package nginx

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// GetPHPSocket returns the PHP-FPM socket path for a given PHP version
func GetPHPSocket(phpVersion string) string {
	if phpVersion == "" {
		phpVersion = "8.2" // Default
	}
	return fmt.Sprintf("/var/run/php/php%s-fpm.sock", phpVersion)
}

// GetDocumentRoot determines the document root for a site
// Looks for common directories: public, public_html, web, or uses site path
func GetDocumentRoot(sitePath string) string {
	// Common Laravel/Symfony/modern PHP structure
	publicDirs := []string{"public", "public_html", "web", "htdocs"}

	for _, dir := range publicDirs {
		fullPath := filepath.Join(sitePath, dir)
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			return fullPath
		}
	}

	// No common directory found, use site path itself
	return sitePath
}

// GenerateConfig generates nginx configuration from a SiteConfig
func GenerateConfig(cfg *SiteConfig) (string, error) {
	tmpl, err := template.New("nginx").Parse(GetTemplate())
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// CreateSiteConfig creates a SiteConfig from basic site information
func CreateSiteConfig(siteName, sitePath, domain, phpVersion string, useSSL bool) *SiteConfig {
	if phpVersion == "" {
		phpVersion = "8.2" // Default
	}

	serverName := fmt.Sprintf("%s.%s", siteName, domain)
	documentRoot := GetDocumentRoot(sitePath)
	phpSocket := GetPHPSocket(phpVersion)

	cfg := &SiteConfig{
		SiteName:   siteName,
		Domain:     domain,
		ServerName: serverName,
		Root:       documentRoot,
		SitePath:   sitePath,
		PHPVersion: phpVersion,
		PHPSocket:  phpSocket,
		UseSSL:     useSSL,
		ListenPort: 80,
	}

	if useSSL {
		certDir := fmt.Sprintf("/home/%s/.phppark/certificates", os.Getenv("USER"))
		cfg.CertPath = filepath.Join(certDir, fmt.Sprintf("%s.crt", siteName))
		cfg.KeyPath = filepath.Join(certDir, fmt.Sprintf("%s.key", siteName))
	}

	return cfg
}

// WriteConfigFile writes the nginx config to a file
func WriteConfigFile(configPath string, content string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

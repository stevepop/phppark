package dns

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	phpParkDnsmasqConf       = "/etc/dnsmasq.d/phppark.conf"
	resolvedConf             = "/etc/systemd/resolved.conf"
	systemdResolveResolvConf = "/run/systemd/resolve/resolv.conf"
	resolvedStubSymlink      = "/run/systemd/resolve/stub-resolv.conf"
)

// SetupDNS configures DNS resolution for .test domains
func SetupDNS(domain string) error {
	return setupLinuxDNS(domain)
}

// RemoveDNS removes DNS configuration for .test domains
func RemoveDNS(domain string) error {
	return removeLinuxDNS(domain)
}

// CheckDNS verifies if DNS is configured
func CheckDNS(domain string) (bool, error) {
	return checkLinuxDNS(domain)
}

// === Linux DNS Setup (dnsmasq) ===

func setupLinuxDNS(domain string) error {
	// Check if dnsmasq is installed
	if _, err := exec.LookPath("dnsmasq"); err != nil {
		return fmt.Errorf("dnsmasq not installed. Install with: sudo apt install dnsmasq")
	}

	// Create dnsmasq domain config
	configPath := fmt.Sprintf("/etc/dnsmasq.d/%s", domain)
	content := fmt.Sprintf("address=/.%s/127.0.0.1\n", domain)

	// Write config (requires sudo)
	cmd := exec.Command("sudo", "tee", configPath)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create dnsmasq config: %w", err)
	}

	// Restart dnsmasq
	if err := exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run(); err != nil {
		return fmt.Errorf("failed to restart dnsmasq: %w", err)
	}

	return nil
}

func removeLinuxDNS(domain string) error {
	configPath := fmt.Sprintf("/etc/dnsmasq.d/%s", domain)

	if err := exec.Command("sudo", "rm", "-f", configPath).Run(); err != nil {
		return fmt.Errorf("failed to remove dnsmasq config: %w", err)
	}

	// If PHPark previously disabled the systemd-resolved stub, revert it
	if IsSystemdResolvedStubDisabled() {
		if err := RevertSystemdResolvedStub(); err != nil {
			fmt.Printf("   ⚠️  Warning: could not revert systemd-resolved stub: %v\n", err)
			fmt.Println("   You may want to manually run: sudo systemctl restart systemd-resolved")
		}
	}

	// Restart dnsmasq if it's running
	exec.Command("sudo", "systemctl", "restart", "dnsmasq").Run()

	return nil
}

func checkLinuxDNS(domain string) (bool, error) {
	configPath := fmt.Sprintf("/etc/dnsmasq.d/%s", domain)
	_, err := os.Stat(configPath)
	return err == nil, nil
}

// === systemd-resolved stub listener management ===

// CheckSystemdResolvedConflict returns true if systemd-resolved is active and
// will conflict with dnsmasq on port 53.
func CheckSystemdResolvedConflict() bool {
	cmd := exec.Command("systemctl", "is-active", "systemd-resolved")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "active"
}

// IsSystemdResolvedStubDisabled returns true if PHPark has previously disabled
// the stub listener (indicated by the presence of phppark.conf).
func IsSystemdResolvedStubDisabled() bool {
	_, err := os.Stat(phpParkDnsmasqConf)
	return err == nil
}

// DisableSystemdResolvedStub disables only the DNS stub listener in systemd-resolved,
// freeing port 53 for dnsmasq while keeping systemd-resolved running so that
// DHCP-provided DNS, VPN routing, and NetworkManager integration all continue to work.
//
// DNS chain after this call:
//
//	/etc/resolv.conf (127.0.0.1) → dnsmasq
//	dnsmasq: *.test  → 127.0.0.1  (handled locally)
//	dnsmasq: all else → /run/systemd/resolve/resolv.conf (live upstream list from systemd-resolved)
func DisableSystemdResolvedStub() error {
	// 1. Set DNSStubListener=no in /etc/systemd/resolved.conf
	if err := setDNSStubListener("no"); err != nil {
		return fmt.Errorf("failed to configure systemd-resolved: %w", err)
	}

	// 2. Restart (not stop/disable) systemd-resolved so it re-reads the config.
	//    It continues running and managing upstream DNS for DHCP/VPN/NetworkManager.
	if err := exec.Command("sudo", "systemctl", "restart", "systemd-resolved").Run(); err != nil {
		return fmt.Errorf("failed to restart systemd-resolved: %w", err)
	}

	// 3. Write /etc/dnsmasq.d/phppark.conf pointing dnsmasq at systemd-resolved's
	//    live upstream file. This prevents a loop: without this, dnsmasq would read
	//    /etc/resolv.conf (which we're about to set to 127.0.0.1) and forward to itself.
	upstreamConf := buildDnsmasqUpstreamConf()
	cmd := exec.Command("sudo", "tee", phpParkDnsmasqConf)
	cmd.Stdin = strings.NewReader(upstreamConf)
	cmd.Stdout = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write dnsmasq upstream config: %w", err)
	}

	// 4. Replace the systemd stub symlink at /etc/resolv.conf with a plain file
	//    pointing to dnsmasq (127.0.0.1). All system DNS queries now go through
	//    dnsmasq, which handles .test locally and forwards everything else upstream.
	info, err := os.Lstat("/etc/resolv.conf")
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		target, _ := os.Readlink("/etc/resolv.conf")
		if strings.Contains(target, "systemd") {
			content := "# Managed by PHPark\nnameserver 127.0.0.1\n"
			cmd = exec.Command("sudo", "tee", "/etc/resolv.conf")
			cmd.Stdin = strings.NewReader(content)
			cmd.Stdout = io.Discard
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update /etc/resolv.conf: %w", err)
			}
		}
	}

	return nil
}

// RevertSystemdResolvedStub reverses the changes made by DisableSystemdResolvedStub:
// it re-enables the stub listener, restores /etc/resolv.conf, and removes phppark.conf.
func RevertSystemdResolvedStub() error {
	// 1. Remove the DNSStubListener=no line from resolved.conf
	if err := setDNSStubListener(""); err != nil {
		return fmt.Errorf("failed to revert resolved.conf: %w", err)
	}

	// 2. Restart systemd-resolved to re-enable the stub listener on 127.0.0.53:53
	if err := exec.Command("sudo", "systemctl", "restart", "systemd-resolved").Run(); err != nil {
		return fmt.Errorf("failed to restart systemd-resolved: %w", err)
	}

	// 3. Remove PHPark's dnsmasq upstream config
	exec.Command("sudo", "rm", "-f", phpParkDnsmasqConf).Run()

	// 4. Restore /etc/resolv.conf to the standard systemd stub symlink
	exec.Command("sudo", "rm", "-f", "/etc/resolv.conf").Run()
	if err := exec.Command("sudo", "ln", "-sf", resolvedStubSymlink, "/etc/resolv.conf").Run(); err != nil {
		return fmt.Errorf("failed to restore /etc/resolv.conf: %w", err)
	}

	return nil
}

// buildDnsmasqUpstreamConf returns the content for /etc/dnsmasq.d/phppark.conf.
// Uses systemd-resolved's live resolver file as upstream when available so that
// VPN, DHCP, and NetworkManager DNS changes are automatically picked up.
// Falls back to public DNS if the file is not yet available.
func buildDnsmasqUpstreamConf() string {
	if _, err := os.Stat(systemdResolveResolvConf); err == nil {
		return fmt.Sprintf("# Managed by PHPark\nresolv-file=%s\n", systemdResolveResolvConf)
	}
	return "# Managed by PHPark\nserver=8.8.8.8\nserver=1.1.1.1\n"
}

// setDNSStubListener writes or removes the DNSStubListener setting in
// /etc/systemd/resolved.conf. Pass "no" to disable the stub; pass "" to
// remove any existing DNSStubListener entry.
func setDNSStubListener(value string) error {
	// Read existing config — file may not exist on minimal installations
	content := ""
	data, err := os.ReadFile(resolvedConf)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", resolvedConf, err)
	}
	if err == nil {
		content = string(data)
	}

	if value == "" {
		// Remove any DNSStubListener line
		lines := strings.Split(content, "\n")
		filtered := lines[:0]
		for _, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "DNSStubListener=") {
				filtered = append(filtered, line)
			}
		}
		content = strings.Join(filtered, "\n")
	} else {
		newSetting := fmt.Sprintf("DNSStubListener=%s", value)
		switch {
		case strings.Contains(content, "DNSStubListener="):
			// Replace the existing setting
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "DNSStubListener=") {
					lines[i] = newSetting
				}
			}
			content = strings.Join(lines, "\n")
		case strings.Contains(content, "[Resolve]"):
			// Inject directly after the [Resolve] heading
			content = strings.Replace(content, "[Resolve]", "[Resolve]\n"+newSetting, 1)
		default:
			// Append a new [Resolve] section
			if content != "" && !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n[Resolve]\n" + newSetting + "\n"
		}
	}

	cmd := exec.Command("sudo", "tee", resolvedConf)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write %s: %w", resolvedConf, err)
	}
	return nil
}

// TestDNSResolution tests if a domain resolves correctly
func TestDNSResolution(hostname string) (bool, error) {
	// Use nslookup to test
	cmd := exec.Command("nslookup", hostname)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil // Domain doesn't resolve
	}

	// Check if it resolves to 127.0.0.1
	outputStr := string(output)
	return strings.Contains(outputStr, "127.0.0.1"), nil
}

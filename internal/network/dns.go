package network

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"kryptx/internal/utils"
)

type DNSManager struct {
	logger      *utils.Logger
	vpnDNS      []string
	originalDNS []string
	configured  bool
}

func NewDNSManager(vpnDNS []string, logger *utils.Logger) *DNSManager {
	return &DNSManager{
		logger: logger,
		vpnDNS: vpnDNS,
	}
}

func (d *DNSManager) Configure() error {
	if d.configured {
		return nil
	}

	d.logger.Info("Configuring DNS for leak protection...")

	// Backup current DNS settings
	if err := d.backupDNS(); err != nil {
		return fmt.Errorf("backing up DNS: %w", err)
	}

	// Set VPN DNS
	if err := d.setVPNDNS(); err != nil {
		return fmt.Errorf("setting VPN DNS: %w", err)
	}

	d.configured = true
	return nil
}

func (d *DNSManager) Restore() error {
	if !d.configured {
		return nil
	}

	d.logger.Info("Restoring original DNS settings...")

	if err := d.restoreDNS(); err != nil {
		return fmt.Errorf("restoring DNS: %w", err)
	}

	d.configured = false
	return nil
}

func (d *DNSManager) backupDNS() error {
	switch runtime.GOOS {
	case "linux":
		return d.backupLinuxDNS()
	case "darwin":
		return d.backupMacOSDNS()
	case "windows":
		return d.backupWindowsDNS()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (d *DNSManager) setVPNDNS() error {
	switch runtime.GOOS {
	case "linux":
		return d.setLinuxDNS()
	case "darwin":
		return d.setMacOSDNS()
	case "windows":
		return d.setWindowsDNS()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (d *DNSManager) restoreDNS() error {
	switch runtime.GOOS {
	case "linux":
		return d.restoreLinuxDNS()
	case "darwin":
		return d.restoreMacOSDNS()
	case "windows":
		return d.restoreWindowsDNS()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func (d *DNSManager) backupLinuxDNS() error {
	// Read current resolv.conf
	data, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err
	}

	// Extract nameservers
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "nameserver") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				d.originalDNS = append(d.originalDNS, parts[1])
			}
		}
	}

	// Backup the file
	return exec.Command("sudo", "cp", "/etc/resolv.conf", "/etc/resolv.conf.kryptx.backup").Run()
}

func (d *DNSManager) setLinuxDNS() error {
	// Create new resolv.conf with VPN DNS
	content := "# KryptX VPN DNS\n"
	for _, dns := range d.vpnDNS {
		content += fmt.Sprintf("nameserver %s\n", dns)
	}

	// Write to temporary file and move
	tmpFile := "/tmp/resolv.conf.kryptx"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return err
	}

	return exec.Command("sudo", "mv", tmpFile, "/etc/resolv.conf").Run()
}

func (d *DNSManager) restoreLinuxDNS() error {
	return exec.Command("sudo", "mv", "/etc/resolv.conf.kryptx.backup", "/etc/resolv.conf").Run()
}

func (d *DNSManager) backupMacOSDNS() error {
	// Get current DNS servers
	cmd := exec.Command("networksetup", "-getdnsservers", "Wi-Fi")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	d.originalDNS = strings.Fields(string(output))
	return nil
}

func (d *DNSManager) setMacOSDNS() error {
	args := []string{"-setdnsservers", "Wi-Fi"}
	args = append(args, d.vpnDNS...)

	cmd := exec.Command("networksetup", args...)
	return cmd.Run()
}

func (d *DNSManager) restoreMacOSDNS() error {
	if len(d.originalDNS) == 0 {
		// Set to empty (automatic)
		return exec.Command("networksetup", "-setdnsservers", "Wi-Fi", "empty").Run()
	}

	args := []string{"-setdnsservers", "Wi-Fi"}
	args = append(args, d.originalDNS...)

	cmd := exec.Command("networksetup", args...)
	return cmd.Run()
}

func (d *DNSManager) backupWindowsDNS() error {
	// Get current DNS servers using netsh
	cmd := exec.Command("netsh", "interface", "ipv4", "show", "dnsservers")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Parse output to extract DNS servers
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Statically Configured DNS Servers:") {
			// Extract DNS servers from following lines
			// This is simplified - real implementation would be more robust
		}
	}

	return nil
}

func (d *DNSManager) setWindowsDNS() error {
	// Set primary DNS
	if len(d.vpnDNS) > 0 {
		cmd := exec.Command("netsh", "interface", "ipv4", "set", "dns", "name=\"Local Area Connection\"", "source=static", "addr="+d.vpnDNS[0])
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Set secondary DNS
	if len(d.vpnDNS) > 1 {
		cmd := exec.Command("netsh", "interface", "ipv4", "add", "dns", "name=\"Local Area Connection\"", "addr="+d.vpnDNS[1])
		return cmd.Run()
	}

	return nil
}

func (d *DNSManager) restoreWindowsDNS() error {
	// Restore to automatic DNS
	cmd := exec.Command("netsh", "interface", "ipv4", "set", "dns", "name=\"Local Area Connection\"", "source=dhcp")
	return cmd.Run()
}

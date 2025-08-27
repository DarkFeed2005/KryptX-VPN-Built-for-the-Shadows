package network

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"kryptx/internal/utils"
)

type KillSwitch struct {
	logger      *utils.Logger
	active      bool
	backupRules []string
}

func NewKillSwitch(logger *utils.Logger) *KillSwitch {
	return &KillSwitch{
		logger: logger,
	}
}

func (k *KillSwitch) Activate() error {
	if k.active {
		return nil
	}

	k.logger.Info("Activating kill switch...")

	switch runtime.GOOS {
	case "linux":
		return k.activateLinux()
	case "darwin":
		return k.activateMacOS()
	case "windows":
		return k.activateWindows()
	default:
		return fmt.Errorf("unsupported OS for kill switch: %s", runtime.GOOS)
	}
}

func (k *KillSwitch) Deactivate() error {
	if !k.active {
		return nil
	}

	k.logger.Info("Deactivating kill switch...")

	switch runtime.GOOS {
	case "linux":
		return k.deactivateLinux()
	case "darwin":
		return k.deactivateMacOS()
	case "windows":
		return k.deactivateWindows()
	default:
		return fmt.Errorf("unsupported OS for kill switch: %s", runtime.GOOS)
	}
}

func (k *KillSwitch) activateLinux() error {
	// Block all outbound traffic except VPN
	rules := []string{
		"iptables -I OUTPUT -o lo -j ACCEPT",
		"iptables -I OUTPUT -o kryptx+ -j ACCEPT",
		"iptables -I OUTPUT -j DROP",
		"ip6tables -I OUTPUT -o lo -j ACCEPT",
		"ip6tables -I OUTPUT -o kryptx+ -j ACCEPT",
		"ip6tables -I OUTPUT -j DROP",
	}

	for _, rule := range rules {
		cmd := exec.Command("sudo", strings.Fields(rule)...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply rule %s: %w", rule, err)
		}
	}

	k.active = true
	return nil
}

func (k *KillSwitch) deactivateLinux() error {
	// Remove kill switch rules
	rules := []string{
		"iptables -D OUTPUT -j DROP",
		"iptables -D OUTPUT -o kryptx+ -j ACCEPT",
		"iptables -D OUTPUT -o lo -j ACCEPT",
		"ip6tables -D OUTPUT -j DROP",
		"ip6tables -D OUTPUT -o kryptx+ -j ACCEPT",
		"ip6tables -D OUTPUT -o lo -j ACCEPT",
	}

	for _, rule := range rules {
		cmd := exec.Command("sudo", strings.Fields(rule)...)
		cmd.Run() // Ignore errors during cleanup
	}

	k.active = false
	return nil
}

func (k *KillSwitch) activateMacOS() error {
	// macOS implementation using pfctl
	pfConfig := `
block out all
pass out on lo0 all
pass out on kryptx0 all
`

	cmd := exec.Command("sudo", "pfctl", "-f", "-")
	cmd.Stdin = strings.NewReader(pfConfig)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to configure pfctl: %w", err)
	}

	// Enable pf
	cmd = exec.Command("sudo", "pfctl", "-e")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable pfctl: %w", err)
	}

	k.active = true
	return nil
}

func (k *KillSwitch) deactivateMacOS() error {
	// Disable pf
	cmd := exec.Command("sudo", "pfctl", "-d")
	cmd.Run() // Ignore errors

	k.active = false
	return nil
}

func (k *KillSwitch) activateWindows() error {
	// Windows implementation using netsh
	rules := []string{
		`netsh advfirewall firewall add rule name="KryptX_Block_All" dir=out action=block`,
		`netsh advfirewall firewall add rule name="KryptX_Allow_Loopback" dir=out action=allow localip=127.0.0.1`,
		`netsh advfirewall firewall add rule name="KryptX_Allow_VPN" dir=out action=allow interface="kryptx"`,
	}

	for _, rule := range rules {
		cmd := exec.Command("cmd", "/C", rule)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to apply rule: %w", err)
		}
	}

	k.active = true
	return nil
}

func (k *KillSwitch) deactivateWindows() error {
	// Remove Windows firewall rules
	rules := []string{
		`netsh advfirewall firewall delete rule name="KryptX_Block_All"`,
		`netsh advfirewall firewall delete rule name="KryptX_Allow_Loopback"`,
		`netsh advfirewall firewall delete rule name="KryptX_Allow_VPN"`,
	}

	for _, rule := range rules {
		cmd := exec.Command("cmd", "/C", rule)
		cmd.Run() // Ignore errors during cleanup
	}

	k.active = false
	return nil
}

func (k *KillSwitch) IsActive() bool {
	return k.active
}

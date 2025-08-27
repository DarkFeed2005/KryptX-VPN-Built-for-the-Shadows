package network

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"kryptx/internal/config"
	"kryptx/internal/utils"
)

type VPNClient struct {
	config     *config.Config
	logger     *utils.Logger
	connected  bool
	killSwitch *KillSwitch
	dnsManager *DNSManager
}

func NewVPNClient(cfg *config.Config, logger *utils.Logger) (*VPNClient, error) {
	client := &VPNClient{
		config: cfg,
		logger: logger,
	}

	if cfg.Security.KillSwitch {
		client.killSwitch = NewKillSwitch(logger)
	}

	if cfg.Security.DNSLeak {
		client.dnsManager = NewDNSManager(cfg.Network.DNS, logger)
	}

	return client, nil
}

func (v *VPNClient) Connect(ctx context.Context) error {
	if v.connected {
		return fmt.Errorf("already connected")
	}

	v.logger.Info("Establishing VPN connection...")

	// Activate kill switch first
	if v.killSwitch != nil {
		if err := v.killSwitch.Activate(); err != nil {
			return fmt.Errorf("activating kill switch: %w", err)
		}
	}

	// Configure DNS
	if v.dnsManager != nil {
		if err := v.dnsManager.Configure(); err != nil {
			return fmt.Errorf("configuring DNS: %w", err)
		}
	}

	// Create WireGuard configuration
	configContent := v.generateWireGuardConfig()

	// Apply configuration based on OS
	if err := v.applyWireGuardConfig(configContent); err != nil {
		return fmt.Errorf("applying WireGuard config: %w", err)
	}

	v.connected = true
	v.logger.Info("VPN connection established")

	// Monitor connection
	go v.monitorConnection(ctx)

	return nil
}

func (v *VPNClient) Disconnect() error {
	if !v.connected {
		return nil
	}

	v.logger.Info("Disconnecting VPN...")

	// Remove WireGuard interface
	if err := v.removeWireGuardInterface(); err != nil {
		v.logger.Error("Failed to remove WireGuard interface: %v", err)
	}

	// Restore DNS
	if v.dnsManager != nil {
		if err := v.dnsManager.Restore(); err != nil {
			v.logger.Error("Failed to restore DNS: %v", err)
		}
	}

	// Deactivate kill switch
	if v.killSwitch != nil {
		if err := v.killSwitch.Deactivate(); err != nil {
			v.logger.Error("Failed to deactivate kill switch: %v", err)
		}
	}

	v.connected = false
	v.logger.Info("VPN disconnected")
	return nil
}

func (v *VPNClient) IsConnected() bool {
	return v.connected
}

func (v *VPNClient) GetStatus() map[string]interface{} {
	status := map[string]interface{}{
		"connected": v.connected,
		"server":    v.config.Server.Endpoint,
	}

	if v.connected {
		// Get connection stats
		if stats := v.getConnectionStats(); stats != nil {
			status["stats"] = stats
		}

		// Get public IP
		if ip := v.getPublicIP(); ip != "" {
			status["public_ip"] = ip
		}
	}

	return status
}

func (v *VPNClient) generateWireGuardConfig() string {
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
MTU = %d

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = %s
PersistentKeepalive = 25
`,
		v.config.Network.PrivateKey,
		v.config.Network.Address,
		strings.Join(v.config.Network.DNS, ", "),
		v.config.Network.MTU,
		v.config.Server.PublicKey,
		v.config.Server.Endpoint,
		v.config.Server.Port,
		strings.Join(v.config.Network.AllowedIPs, ", "),
	)
}

func (v *VPNClient) applyWireGuardConfig(configContent string) error {
	switch runtime.GOOS {
	case "windows":
		return v.applyWindowsConfig(configContent)
	case "darwin":
		return v.applyMacOSConfig(configContent)
	case "linux":
		return v.applyLinuxConfig(configContent)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (v *VPNClient) applyLinuxConfig(configContent string) error {
	interfaceName := v.config.Network.Interface

	// Create WireGuard interface
	cmd := exec.Command("sudo", "ip", "link", "add", "dev", interfaceName, "type", "wireguard")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("creating interface: %w", err)
	}

	// Apply configuration using wg-quick
	cmd = exec.Command("sudo", "wg-quick", "up", "/dev/stdin")
	cmd.Stdin = strings.NewReader(configContent)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("applying config: %w", err)
	}

	return nil
}

func (v *VPNClient) applyWindowsConfig(configContent string) error {
	// For Windows, we would typically use the WireGuard service API
	// This is a simplified version - real implementation would use WG API
	return fmt.Errorf("Windows implementation requires WireGuard service integration")
}

func (v *VPNClient) applyMacOSConfig(configContent string) error {
	// Similar to Linux but may require different commands
	return v.applyLinuxConfig(configContent)
}

func (v *VPNClient) removeWireGuardInterface() error {
	interfaceName := v.config.Network.Interface

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("sudo", "wg-quick", "down", interfaceName)
		return cmd.Run()
	case "windows":
		// Windows-specific cleanup
		return nil
	default:
		return fmt.Errorf("unsupported OS")
	}
}

func (v *VPNClient) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !v.isInterfaceUp() {
				v.logger.Warning("Connection lost, attempting reconnect...")
				v.connected = false
				// Attempt reconnection logic here
			}
		}
	}
}

func (v *VPNClient) isInterfaceUp() bool {
	_, err := net.InterfaceByName(v.config.Network.Interface)
	return err == nil
}

func (v *VPNClient) getConnectionStats() map[string]interface{} {
	// Implementation would parse `wg show` output
	return map[string]interface{}{
		"bytes_sent":     0,
		"bytes_received": 0,
		"last_handshake": time.Now(),
	}
}

func (v *VPNClient) getPublicIP() string {
	// Simple HTTP request to get public IP
	// Implementation would make HTTP request to IP service
	return "0.0.0.0"
}

package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Network  NetworkConfig  `yaml:"network"`
	Security SecurityConfig `yaml:"security"`
	GUI      GUIConfig      `yaml:"gui"`
}

type ServerConfig struct {
	Endpoint  string `yaml:"endpoint"`
	PublicKey string `yaml:"public_key"`
	Port      int    `yaml:"port"`
}

type NetworkConfig struct {
	Interface  string   `yaml:"interface"`
	PrivateKey string   `yaml:"private_key"`
	Address    string   `yaml:"address"`
	DNS        []string `yaml:"dns"`
	AllowedIPs []string `yaml:"allowed_ips"`
	MTU        int      `yaml:"mtu"`
}

type SecurityConfig struct {
	KillSwitch    bool   `yaml:"kill_switch"`
	DNSLeak       bool   `yaml:"dns_leak_protection"`
	EncryptConfig bool   `yaml:"encrypt_config"`
	VaultPassword string `yaml:"vault_password"`
}

type GUIConfig struct {
	Theme       string `yaml:"theme"`
	Animated    bool   `yaml:"animated"`
	StartHidden bool   `yaml:"start_hidden"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Generate private key if not present
	if config.Network.PrivateKey == "" {
		key, err := generatePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("generating private key: %w", err)
		}
		config.Network.PrivateKey = key
	}

	return &config, nil
}

func generatePrivateKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

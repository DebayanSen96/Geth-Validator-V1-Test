package p2p

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// DefaultP2PConfig returns the default p2p configuration.
func DefaultP2PConfig() Config {
	return Config{
		ListenAddresses: []string{
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/udp/9000/quic-v1",
		},
		BootstrapPeers: []string{},
		PrivateKeyFile: "",
	}
}

// LoadP2PConfig loads the p2p configuration from a file.
func LoadP2PConfig(dataDir string) (Config, error) {
	configPath := filepath.Join(dataDir, "p2p_config.json")
	config := DefaultP2PConfig()

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		if err := SaveP2PConfig(dataDir, config); err != nil {
			return config, fmt.Errorf("failed to save default p2p config: %w", err)
		}
		return config, nil
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read p2p config: %w", err)
	}

	// Parse the config
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse p2p config: %w", err)
	}

	return config, nil
}

// SaveP2PConfig saves the p2p configuration to a file.
func SaveP2PConfig(dataDir string, config Config) error {
	// Create the data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Marshal the config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal p2p config: %w", err)
	}

	// Write the config file
	configPath := filepath.Join(dataDir, "p2p_config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write p2p config: %w", err)
	}

	return nil
}

package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	DefaultLogDir         = "logs"
	DefaultViewer         = "less"
	DefaultAutoPruneHours = 24 // auto-prune done jobs older than 24hrs (0 = disabled)
)

type Config struct {
	LogDir         string `toml:"log_dir"`
	Viewer         string `toml:"viewer"`
	AutoPruneHours int    `toml:"auto_prune_hours"` // auto-clear done jobs older than N hours (0 = disabled)
}

// ConfigDir returns the bj config directory path
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "bj"), nil
}

// DefaultConfig returns a Config with default values
func DefaultConfig() Config {
	return Config{
		LogDir:         DefaultLogDir,
		Viewer:         DefaultViewer,
		AutoPruneHours: DefaultAutoPruneHours,
	}
}

// Load reads the config file, creating it with defaults if it doesn't exist
func Load() (*Config, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "bj.toml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create config directory and default config
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, err
		}

		cfg := DefaultConfig()
		if err := Save(&cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}

	// Load existing config
	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults for missing fields
	if cfg.LogDir == "" {
		cfg.LogDir = DefaultLogDir
	}
	if cfg.Viewer == "" {
		cfg.Viewer = DefaultViewer
	}

	return &cfg, nil
}

// Save writes the config to disk
func Save(cfg *Config) error {
	configDir, err := ConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "bj.toml")

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}

// LogDir returns the absolute path to the log directory
func (c *Config) LogDirPath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	// If LogDir is relative, make it relative to config dir
	if !filepath.IsAbs(c.LogDir) {
		return filepath.Join(configDir, c.LogDir), nil
	}
	return c.LogDir, nil
}

// EnsureLogDir creates the log directory if it doesn't exist
func (c *Config) EnsureLogDir() error {
	logDir, err := c.LogDirPath()
	if err != nil {
		return err
	}
	return os.MkdirAll(logDir, 0755)
}

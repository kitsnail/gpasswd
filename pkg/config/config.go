package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Session struct {
		Timeout int `mapstructure:"timeout"` // seconds, 0 = no timeout
	} `mapstructure:"session"`

	Clipboard struct {
		ClearTimeout int `mapstructure:"clear_timeout"` // seconds
	} `mapstructure:"clipboard"`

	PasswordGenerator struct {
		Length           int  `mapstructure:"length"`
		UseUppercase     bool `mapstructure:"use_uppercase"`
		UseLowercase     bool `mapstructure:"use_lowercase"`
		UseDigits        bool `mapstructure:"use_digits"`
		UseSymbols       bool `mapstructure:"use_symbols"`
		ExcludeAmbiguous bool `mapstructure:"exclude_ambiguous"`
	} `mapstructure:"password_generator"`

	Security struct {
		FailedAttemptsLimit int `mapstructure:"failed_attempts_limit"`
		LockoutDuration     int `mapstructure:"lockout_duration"` // seconds
	} `mapstructure:"security"`

	Argon2 struct {
		TimeCost    uint32 `mapstructure:"time_cost"`
		MemoryCost  uint32 `mapstructure:"memory_cost"` // KB
		Parallelism uint8  `mapstructure:"parallelism"`
	} `mapstructure:"argon2"`

	Display struct {
		ShowTimestamps bool   `mapstructure:"show_timestamps"`
		DateFormat     string `mapstructure:"date_format"`
	} `mapstructure:"display"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	cfg := &Config{}

	cfg.Session.Timeout = 300 // 5 minutes

	cfg.Clipboard.ClearTimeout = 30

	cfg.PasswordGenerator.Length = 20
	cfg.PasswordGenerator.UseUppercase = true
	cfg.PasswordGenerator.UseLowercase = true
	cfg.PasswordGenerator.UseDigits = true
	cfg.PasswordGenerator.UseSymbols = true
	cfg.PasswordGenerator.ExcludeAmbiguous = false

	cfg.Security.FailedAttemptsLimit = 5
	cfg.Security.LockoutDuration = 30

	cfg.Argon2.TimeCost = 3
	cfg.Argon2.MemoryCost = 65536 // 64 MB
	cfg.Argon2.Parallelism = 4

	cfg.Display.ShowTimestamps = true
	cfg.Display.DateFormat = "2006-01-02 15:04"

	return cfg
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get home directory: %v", err))
	}
	return filepath.Join(home, ".gpasswd")
}

// GetVaultPath returns the path to the vault database
func GetVaultPath() string {
	return filepath.Join(GetConfigDir(), "vault.db")
}

// Load loads the configuration from the config file
func Load() (*Config, error) {
	configDir := GetConfigDir()
	configFile := filepath.Join(configDir, "config.yaml")

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	cfg := DefaultConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to the config file
func (c *Config) Save() error {
	configDir := GetConfigDir()
	configFile := filepath.Join(configDir, "config.yaml")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// Marshal config to viper
	viper.Set("session", c.Session)
	viper.Set("clipboard", c.Clipboard)
	viper.Set("password_generator", c.PasswordGenerator)
	viper.Set("security", c.Security)
	viper.Set("argon2", c.Argon2)
	viper.Set("display", c.Display)

	if err := viper.WriteConfig(); err != nil {
		// If config file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}
			return nil
		}
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Rule struct {
	Name        string   `yaml:"name,omitempty"`
	Extensions  []string `yaml:"extensions,omitempty"`
	Regex       string   `yaml:"regex,omitempty" validate:"omitempty,is-regex"`
	Mime        string   `yaml:"mime,omitempty" validate:"omitempty,is-regex"`
	Scheme      string   `yaml:"scheme,omitempty"`
	OS          []string `yaml:"os,omitempty"`
	Background  bool     `yaml:"background,omitempty"`
	Terminal    bool     `yaml:"terminal,omitempty"`
	Fallthrough bool     `yaml:"fallthrough,omitempty"`
	Command     string            `yaml:"command,omitempty" validate:"required"`
	Script      string            `yaml:"script,omitempty"` // JavaScript code
	Env         map[string]string `yaml:"env,omitempty"`    // Environment variables
}

type Config struct {
	Version        string            `yaml:"version"`
	DefaultCommand string            `yaml:"default_command,omitempty"`
	Default        string            `yaml:"default,omitempty"` // Shorter alias for DefaultCommand
	Aliases        map[string]string `yaml:"aliases,omitempty"`
	Rules          []Rule            `yaml:"rules" validate:"dive"`
	Sync           *SyncConfig       `yaml:"sync,omitempty"`
}

type SyncConfig struct {
	GistID string `yaml:"gist_id,omitempty"`
	Token  string `yaml:"token,omitempty"` // Optional: usually passed via env var or flag, but can be stored
}

func LoadConfig(path string) (*Config, error) {
	configPath, err := GetConfigPath(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist, or maybe error?
			// For now, let's return an error so the user knows they need a config.
			return nil, fmt.Errorf("config file not found at %s", configPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// If 'default' is set, use it as DefaultCommand (unless DefaultCommand is already set)
	if cfg.Default != "" && cfg.DefaultCommand == "" {
		cfg.DefaultCommand = cfg.Default
	}

	return &cfg, nil
}

func GetConfigPath(path string) (string, error) {
	return GetConfigPathWithProfile(path, "")
}

// UserHomeDir is a variable to allow mocking in tests
var UserHomeDir = os.UserHomeDir

// GetConfigPathWithProfile returns the config file path for a specific profile
// If profile is empty, returns the default config path
func GetConfigPathWithProfile(path string, profile string) (string, error) {
	if path != "" {
		return path, nil
	}

	home, err := UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	if profile == "" {
		return filepath.Join(home, ".config", "via", "config.yml"), nil
	}

	return filepath.Join(home, ".config", "via", "profiles", profile+".yml"), nil
}

func SaveConfig(path string, cfg *Config) error {
	configPath, err := GetConfigPath(path)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ValidateConfig(cfg *Config) error {
	validate := validator.New()
	
	// Register custom validation for regex
	validate.RegisterValidation("is-regex", func(fl validator.FieldLevel) bool {
		_, err := regexp.Compile(fl.Field().String())
		return err == nil
	})

	if err := validate.Struct(cfg); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Simplify error message for the user
			for _, e := range validationErrors {
				// e.Namespace() gives full path like Config.Rules[0].Command
				return fmt.Errorf("validation failed: %s is %s", e.Namespace(), e.Tag())
			}
		}
		return err
	}
	return nil
}

// ValidateRegex validates a regex pattern
func ValidateRegex(pattern string) error {
	_, err := regexp.Compile(pattern)
	return err
}

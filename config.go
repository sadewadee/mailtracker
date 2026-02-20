package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds all configuration
type AppConfig struct {
	API_TOKEN           string `json:"api_token"`
	NOTIFY_EMAIL        string `json:"notify_email,omitempty"`
	EXIM_LOG            string `json:"exim_log"`
	WHM_API_HOST        string `json:"whm_api_host"`
	PREFER_MODERN_UAPI  string `json:"prefer_modern_uapi"`
	MAX_PER_MIN         int16  `json:"max_per_min"`
	MAX_PER_HOUR        int16  `json:"max_per_hour"`
	TELEGRAM_BOT_TOKEN  string `json:"telegram_bot_token,omitempty"`
	TELEGRAM_ADMIN_IDS  string `json:"telegram_admin_ids,omitempty"`
	TELEGRAM_NOTIFY_CHAT_ID string `json:"telegram_notify_chat_id,omitempty"`
	SLACK_BOT_TOKEN     string `json:"slack_bot_token,omitempty"`
	SLACK_ADMIN_IDS     string `json:"slack_admin_ids,omitempty"`
	SLACK_NOTIFY_CHANNEL string `json:"slack_notify_channel,omitempty"`
}

var (
	appConfigPath = ".eximmon.conf"
	appConfig     *AppConfig
)

// loadConfig loads config from file, returns nil if not exists
func loadConfig() *AppConfig {
	// Try multiple paths
	paths := []string{
		appConfigPath,
		filepath.Join(os.Getenv("HOME"), ".eximmon.conf"),
		"/etc/eximmon.conf",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cfg AppConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			log("Warning: failed to parse config file %s: %v", path, err)
			continue
		}

		log("Loaded config from: %s", path)
		return &cfg
	}

	return nil
}

// saveConfig saves config to file with secure permissions
func saveConfig(cfg *AppConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Save to current directory
	path := appConfigPath
	if err := os.WriteFile(path, data, 0600); err != nil {
		// Try home directory
		path = filepath.Join(os.Getenv("HOME"), ".eximmon.conf")
		if err := os.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	log("Config saved to: %s (permission: 600)", path)
	return nil
}

// applyConfigToEnv applies config values to environment variables
// Environment variables take precedence over config file
func applyConfigToEnv(cfg *AppConfig) {
	if cfg == nil {
		return
	}

	// Only set from config if env var is not already set
	if os.Getenv("API_TOKEN") == "" && cfg.API_TOKEN != "" {
		os.Setenv("API_TOKEN", cfg.API_TOKEN)
	}
	if os.Getenv("NOTIFY_EMAIL") == "" && cfg.NOTIFY_EMAIL != "" {
		os.Setenv("NOTIFY_EMAIL", cfg.NOTIFY_EMAIL)
	}
	if os.Getenv("EXIM_LOG") == "" && cfg.EXIM_LOG != "" {
		os.Setenv("EXIM_LOG", cfg.EXIM_LOG)
	} else if cfg.EXIM_LOG != "" {
		// Default from config
	} else {
		os.Setenv("EXIM_LOG", "/var/log/exim_mainlog")
	}
	if os.Getenv("WHM_API_HOST") == "" && cfg.WHM_API_HOST != "" {
		os.Setenv("WHM_API_HOST", cfg.WHM_API_HOST)
	}
	if os.Getenv("PREFER_MODERN_UAPI") == "" && cfg.PREFER_MODERN_UAPI != "" {
		os.Setenv("PREFER_MODERN_UAPI", cfg.PREFER_MODERN_UAPI)
	}
	if os.Getenv("MAX_PER_MIN") == "" && cfg.MAX_PER_MIN > 0 {
		os.Setenv("MAX_PER_MIN", fmt.Sprintf("%d", cfg.MAX_PER_MIN))
	}
	if os.Getenv("MAX_PER_HOUR") == "" && cfg.MAX_PER_HOUR > 0 {
		os.Setenv("MAX_PER_HOUR", fmt.Sprintf("%d", cfg.MAX_PER_HOUR))
	}
	if os.Getenv("TELEGRAM_BOT_TOKEN") == "" && cfg.TELEGRAM_BOT_TOKEN != "" {
		os.Setenv("TELEGRAM_BOT_TOKEN", cfg.TELEGRAM_BOT_TOKEN)
	}
	if os.Getenv("TELEGRAM_ADMIN_IDS") == "" && cfg.TELEGRAM_ADMIN_IDS != "" {
		os.Setenv("TELEGRAM_ADMIN_IDS", cfg.TELEGRAM_ADMIN_IDS)
	}
	if os.Getenv("TELEGRAM_NOTIFY_CHAT_ID") == "" && cfg.TELEGRAM_NOTIFY_CHAT_ID != "" {
		os.Setenv("TELEGRAM_NOTIFY_CHAT_ID", cfg.TELEGRAM_NOTIFY_CHAT_ID)
	}
	if os.Getenv("SLACK_BOT_TOKEN") == "" && cfg.SLACK_BOT_TOKEN != "" {
		os.Setenv("SLACK_BOT_TOKEN", cfg.SLACK_BOT_TOKEN)
	}
	if os.Getenv("SLACK_ADMIN_IDS") == "" && cfg.SLACK_ADMIN_IDS != "" {
		os.Setenv("SLACK_ADMIN_IDS", cfg.SLACK_ADMIN_IDS)
	}
	if os.Getenv("SLACK_NOTIFY_CHANNEL") == "" && cfg.SLACK_NOTIFY_CHANNEL != "" {
		os.Setenv("SLACK_NOTIFY_CHANNEL", cfg.SLACK_NOTIFY_CHANNEL)
	}
}

// mergeEnvToConfig merges current env vars into config
func mergeEnvToConfig(cfg *AppConfig) *AppConfig {
	if cfg == nil {
		cfg = &AppConfig{
			EXIM_LOG: "/var/log/exim_mainlog",
			MAX_PER_MIN: 8,
			MAX_PER_HOUR: 100,
		}
	}

	if v := os.Getenv("API_TOKEN"); v != "" {
		cfg.API_TOKEN = v
	}
	if v := os.Getenv("NOTIFY_EMAIL"); v != "" {
		cfg.NOTIFY_EMAIL = v
	}
	if v := os.Getenv("EXIM_LOG"); v != "" {
		cfg.EXIM_LOG = v
	}
	if v := os.Getenv("WHM_API_HOST"); v != "" {
		cfg.WHM_API_HOST = v
	}
	if v := os.Getenv("PREFER_MODERN_UAPI"); v != "" {
		cfg.PREFER_MODERN_UAPI = v
	}
	if v := os.Getenv("MAX_PER_MIN"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.MAX_PER_MIN)
	}
	if v := os.Getenv("MAX_PER_HOUR"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.MAX_PER_HOUR)
	}
	if v := os.Getenv("TELEGRAM_BOT_TOKEN"); v != "" {
		cfg.TELEGRAM_BOT_TOKEN = v
	}
	if v := os.Getenv("TELEGRAM_ADMIN_IDS"); v != "" {
		cfg.TELEGRAM_ADMIN_IDS = v
	}
	if v := os.Getenv("TELEGRAM_NOTIFY_CHAT_ID"); v != "" {
		cfg.TELEGRAM_NOTIFY_CHAT_ID = v
	}
	if v := os.Getenv("SLACK_BOT_TOKEN"); v != "" {
		cfg.SLACK_BOT_TOKEN = v
	}
	if v := os.Getenv("SLACK_ADMIN_IDS"); v != "" {
		cfg.SLACK_ADMIN_IDS = v
	}
	if v := os.Getenv("SLACK_NOTIFY_CHANNEL"); v != "" {
		cfg.SLACK_NOTIFY_CHANNEL = v
	}

	return cfg
}

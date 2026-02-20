package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds CLI configuration.
type Config struct {
	DBPath      string
	YelpAPIKey  string
	YelpEnabled bool
}

// ParseFlags parses command-line flags and returns configuration.
func ParseFlags() (*Config, error) {
	config := &Config{}

	// Load .env files first so env-based defaults work with existing flag parsing.
	loadDotEnv(".env")
	loadDotEnv(".env.local")

	flag.StringVar(&config.DBPath, "db", "", "Path to SQLite database file (default: ~/.toni/toni.db)")
	flag.StringVar(&config.YelpAPIKey, "yelp-key", "", "Yelp Fusion API key (or set YELP_API_KEY env var)")
	flag.Parse()

	// Get Yelp API key from env if not provided via flag
	if config.YelpAPIKey == "" {
		config.YelpAPIKey = os.Getenv("YELP_API_KEY")
	}

	// Set default DB path if not specified
	var configDir string
	if config.DBPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		configDir = filepath.Join(home, ".toni")
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		config.DBPath = filepath.Join(configDir, "toni.db")
	} else {
		configDir = filepath.Dir(config.DBPath)
	}

	settings, err := loadOnboardingSettings(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load onboarding settings: %w", err)
	}

	if shouldRunOnboarding(settings) {
		settings, err = runOnboarding(configDir, config.YelpAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to run onboarding: %w", err)
		}
	}

	config.YelpEnabled = settings.YelpEnabled
	if config.YelpAPIKey == "" && settings.YelpEnabled {
		secureKey, err := loadSecureYelpAPIKey(configDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load secure Yelp API key: %w", err)
		}
		config.YelpAPIKey = strings.TrimSpace(secureKey)
	}
	if config.YelpAPIKey != "" {
		config.YelpEnabled = true
	}

	return config, nil
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		value = strings.Trim(value, `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

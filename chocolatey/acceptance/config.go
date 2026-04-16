// Package acceptance provides shared infrastructure for acceptance tests that
// exercise the go-sdk-chocolatey SDK against a live Chocolatey NuGet v2 feed.
//
// Acceptance tests require network access. They are skipped automatically when
// the CHOCOLATEY_ACCEPTANCE environment variable is not set to "true".
//
// Environment variables:
//
//	CHOCOLATEY_ACCEPTANCE  set to "true" to enable acceptance tests (required)
//	CHOCOLATEY_BASE_URL    NuGet v2 feed root (default: community repository)
//	CHOCOLATEY_API_KEY     X-NuGet-ApiKey header value (optional; community repo needs none)
//	CHOCOLATEY_TIMEOUT     per-request timeout, e.g. "30s" (default: 30s)
//	CHOCOLATEY_VERBOSE     set to "true" for verbose test logging
package acceptance

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/config"
)

// TestConfig holds configuration for acceptance tests driven by environment variables.
type TestConfig struct {
	// BaseURL is the NuGet v2 feed root. Defaults to the community repository.
	BaseURL string

	// APIKey is the X-NuGet-ApiKey header value. Optional for the community repo.
	APIKey string

	// RequestTimeout is the per-request context deadline.
	RequestTimeout time.Duration

	// Verbose enables detailed test logging.
	Verbose bool
}

var (
	// Config is the global acceptance-test configuration, populated from env on init.
	Config *TestConfig

	// Client is the shared SDK client for acceptance tests.
	Client *chocolatey.Client
)

func init() {
	Config = &TestConfig{
		BaseURL:        getEnv("CHOCOLATEY_BASE_URL", ""),
		APIKey:         getEnv("CHOCOLATEY_API_KEY", ""),
		RequestTimeout: getDurationEnv("CHOCOLATEY_TIMEOUT", 30*time.Second),
		Verbose:        getBoolEnv("CHOCOLATEY_VERBOSE", false),
	}
}

// InitClient creates the shared SDK client from the environment configuration.
func InitClient() error {
	cfg := &config.Config{
		BaseURL: Config.BaseURL,
		APIKey:  Config.APIKey,
	}

	opts := []chocolatey.ClientOption{
		chocolatey.WithTimeout(Config.RequestTimeout),
	}

	var err error
	Client, err = chocolatey.NewClient(cfg, opts...)
	if err != nil {
		return fmt.Errorf("failed to create Chocolatey client: %w", err)
	}

	if Config.Verbose {
		baseURL := Config.BaseURL
		if baseURL == "" {
			baseURL = "https://community.chocolatey.org/api/v2 (default)"
		}
		log.Printf("Acceptance test client initialised: %s", baseURL)
	}

	return nil
}

// IsConfigured returns true when acceptance tests have been explicitly opted in.
// We require an explicit opt-in because acceptance tests hit the live network.
func IsConfigured() bool {
	return getBoolEnv("CHOCOLATEY_ACCEPTANCE", false)
}

// ── env helpers ───────────────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	v := os.Getenv(key)
	switch v {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

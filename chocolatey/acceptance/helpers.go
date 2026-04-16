package acceptance

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey"
	"github.com/stretchr/testify/require"
)

// SkipIfNotConfigured skips the test when CHOCOLATEY_ACCEPTANCE is not set to "true".
func SkipIfNotConfigured(t *testing.T) {
	t.Helper()
	if !IsConfigured() {
		t.Skip("CHOCOLATEY_ACCEPTANCE not set to true, skipping acceptance test")
	}
}

// RequireClient ensures the shared client is initialised, skipping the test if
// CHOCOLATEY_ACCEPTANCE is absent or if client construction fails.
func RequireClient(t *testing.T) {
	t.Helper()
	SkipIfNotConfigured(t)

	if Client == nil {
		err := InitClient()
		require.NoError(t, err, "Failed to initialise Chocolatey client")
	}
}

// NewContext creates a context with the configured request timeout.
func NewContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), Config.RequestTimeout)
}

// LogTestStage logs a named test stage with optional GitHub Actions annotation.
func LogTestStage(t *testing.T, stage, message string, args ...any) {
	t.Helper()
	formatted := message
	if len(args) > 0 {
		formatted = fmt.Sprintf(message, args...)
	}
	if isGitHubActions() {
		fmt.Printf("::notice title=%s::%s\n", stage, formatted)
	}
	if Config.Verbose {
		t.Logf("[%s] %s", stage, formatted)
	}
}

// LogTestSuccess logs a successful step.
func LogTestSuccess(t *testing.T, message string, args ...any) {
	t.Helper()
	formatted := message
	if len(args) > 0 {
		formatted = fmt.Sprintf(message, args...)
	}
	if isGitHubActions() {
		fmt.Printf("::notice title=Success::%s\n", formatted)
	}
	if Config.Verbose {
		t.Logf("OK: %s", formatted)
	}
}

// LogTestWarning logs a non-fatal warning.
func LogTestWarning(t *testing.T, message string, args ...any) {
	t.Helper()
	formatted := message
	if len(args) > 0 {
		formatted = fmt.Sprintf(message, args...)
	}
	if isGitHubActions() {
		fmt.Printf("::warning title=Warning::%s\n", formatted)
	}
	if Config.Verbose {
		t.Logf("WARNING: %s", formatted)
	}
}

// RetryOnNotFound retries a call when it returns a 404 error, with exponential
// backoff. Used for eventual consistency on feeds that may lag after writes.
func RetryOnNotFound(t *testing.T, maxRetries int, initialDelay time.Duration, fn func() error) error {
	t.Helper()
	delay := initialDelay
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if !chocolatey.IsNotFound(lastErr) {
			return lastErr
		}
		if i < maxRetries-1 {
			if Config.Verbose {
				t.Logf("404 received, retry %d/%d — waiting %v", i+1, maxRetries, delay)
			}
			time.Sleep(delay)
			delay *= 2
		}
	}
	return lastErr
}

// PollUntil retries fn every interval until it returns true or the timeout elapses.
func PollUntil(t *testing.T, timeout, interval time.Duration, fn func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

func isGitHubActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

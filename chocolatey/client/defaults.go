package client

import (
	"time"

	"github.com/deploymenttheory/go-sdk-chocolatey/chocolatey/constants"
)

const (
	UserAgentBase = constants.UserAgentBase

	DefaultTimeout      = constants.DefaultTimeout
	MaxRetries          = constants.MaxRetries
	RetryWaitTime       = constants.RetryWaitTime
	RetryMaxWaitTime    = constants.RetryMaxWaitTime
	DefaultPageSize     = constants.DefaultPageSize

	// adaptiveDelayMax caps the mandatory request delay to prevent unbounded pauses.
	adaptiveDelayMax = 5 * time.Second
)

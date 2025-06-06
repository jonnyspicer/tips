package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// TestConfig holds configuration for API testing
type TestConfig struct {
	MaxRetries     int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
	RateLimit      time.Duration
	MaxCostUSD     float64 // Maximum cost per test run
}

// DefaultTestConfig returns safe defaults for API testing
func DefaultTestConfig() TestConfig {
	return TestConfig{
		MaxRetries:     3,
		RetryDelay:     time.Second * 2,
		RequestTimeout: time.Second * 30,
		RateLimit:      time.Second * 1,
		MaxCostUSD:     0.10, // 10 cents max per test run
	}
}

// APITestRunner handles API testing with rate limiting and cost control
type APITestRunner struct {
	config       TestConfig
	requestCount int
	totalCost    float64
	lastRequest  time.Time
}

// NewAPITestRunner creates a new test runner with the given config
func NewAPITestRunner(config TestConfig) *APITestRunner {
	return &APITestRunner{
		config: config,
	}
}

// ExecuteWithRetry executes a function with retry logic and rate limiting
func (r *APITestRunner) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	// Rate limiting
	if !r.lastRequest.IsZero() {
		elapsed := time.Since(r.lastRequest)
		if elapsed < r.config.RateLimit {
			time.Sleep(r.config.RateLimit - elapsed)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := r.config.RetryDelay * time.Duration(attempt)
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(delay + jitter)
		}

		// Create timeout context for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, r.config.RequestTimeout)
		_ = attemptCtx // Mark as used to avoid compiler warning

		err := operation()
		cancel()

		if err == nil {
			r.requestCount++
			r.lastRequest = time.Now()
			return nil
		}

		lastErr = err

		// Check if we should retry based on error type
		if !isRetryableError(err) {
			break
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", r.config.MaxRetries, lastErr)
}

// EstimateAndCheckCost estimates the cost of an operation and checks against limits
func (r *APITestRunner) EstimateAndCheckCost(provider string, promptTokens, maxTokens int) error {
	estimatedCost := estimateAPICost(provider, promptTokens, maxTokens)

	if r.totalCost+estimatedCost > r.config.MaxCostUSD {
		return fmt.Errorf("estimated cost (%.4f USD) would exceed limit (%.4f USD)",
			r.totalCost+estimatedCost, r.config.MaxCostUSD)
	}

	r.totalCost += estimatedCost
	return nil
}

// GetStats returns current test statistics
func (r *APITestRunner) GetStats() (requestCount int, totalCost float64) {
	return r.requestCount, r.totalCost
}

// estimateAPICost provides rough cost estimates for API calls
func estimateAPICost(provider string, promptTokens, maxTokens int) float64 {
	// Rough estimates based on current pricing (as of 2024)
	// These should be updated regularly

	totalTokens := promptTokens + maxTokens

	switch provider {
	case "openai":
		// GPT-3.5-turbo pricing: ~$0.002 per 1k tokens
		return float64(totalTokens) / 1000.0 * 0.002
	case "anthropic":
		// Claude pricing: ~$0.008 per 1k tokens
		return float64(totalTokens) / 1000.0 * 0.008
	case "google":
		// Gemini pricing: ~$0.001 per 1k tokens
		return float64(totalTokens) / 1000.0 * 0.001
	default:
		// Conservative estimate
		return float64(totalTokens) / 1000.0 * 0.01
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Common retryable errors
	retryableErrors := []string{
		"rate limit",
		"timeout",
		"temporary failure",
		"connection refused",
		"network error",
		"500", // Internal server error
		"502", // Bad gateway
		"503", // Service unavailable
		"504", // Gateway timeout
	}

	for _, retryable := range retryableErrors {
		if contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

// indexOf returns the index of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

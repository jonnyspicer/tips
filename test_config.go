package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type TestConfig struct {
	MaxRetries     int
	RetryDelay     time.Duration
	RequestTimeout time.Duration
	RateLimit      time.Duration
	MaxCostUSD     float64
}

func DefaultTestConfig() TestConfig {
	return TestConfig{
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,
		RequestTimeout: 30 * time.Second,
		RateLimit:      time.Second,
		MaxCostUSD:     0.10,
	}
}

type APITestRunner struct {
	config       TestConfig
	requestCount int
	totalCost    float64
	lastRequest  time.Time
}

func NewAPITestRunner(config TestConfig) *APITestRunner {
	return &APITestRunner{config: config}
}

func (r *APITestRunner) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	if !r.lastRequest.IsZero() {
		if elapsed := time.Since(r.lastRequest); elapsed < r.config.RateLimit {
			time.Sleep(r.config.RateLimit - elapsed)
		}
	}

	var lastErr error
	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := r.config.RetryDelay * time.Duration(attempt)
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(delay + jitter)
		}

		_, cancel := context.WithTimeout(ctx, r.config.RequestTimeout)
		err := operation()
		cancel()

		if err == nil {
			r.requestCount++
			r.lastRequest = time.Now()
			return nil
		}

		lastErr = err
		if !isRetryableError(err) {
			break
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", r.config.MaxRetries, lastErr)
}

func (r *APITestRunner) EstimateAndCheckCost(provider string, promptTokens, maxTokens int) error {
	estimatedCost := estimateAPICost(provider, promptTokens, maxTokens)
	if r.totalCost+estimatedCost > r.config.MaxCostUSD {
		return fmt.Errorf("estimated cost (%.4f USD) would exceed limit (%.4f USD)",
			r.totalCost+estimatedCost, r.config.MaxCostUSD)
	}
	r.totalCost += estimatedCost
	return nil
}

func (r *APITestRunner) GetStats() (requestCount int, totalCost float64) {
	return r.requestCount, r.totalCost
}

func estimateAPICost(provider string, promptTokens, maxTokens int) float64 {
	totalTokens := float64(promptTokens + maxTokens)
	switch provider {
	case "openai":
		return totalTokens / 1000.0 * 0.002
	case "anthropic":
		return totalTokens / 1000.0 * 0.008
	case "google":
		return totalTokens / 1000.0 * 0.001
	default:
		return totalTokens / 1000.0 * 0.01
	}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	retryableErrors := []string{
		"rate limit", "timeout", "temporary failure", "connection refused",
		"network error", "500", "502", "503", "504",
	}
	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	return false
}

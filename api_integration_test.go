//go:build integration && api
// +build integration,api

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestRealAPIIntegration tests actual API calls with real providers
// Run with: go test -tags="integration,api" -v ./...
// Requires environment variables: OPENAI_API_KEY, ANTHROPIC_API_KEY, or GOOGLE_API_KEY
func TestRealAPIIntegration(t *testing.T) {
	// Skip if no API keys are provided
	if !hasAnyAPIKey() {
		t.Skip("Skipping API integration tests: no API keys found")
	}

	tests := []struct {
		name       string
		provider   string
		envVar     string
		topic      string
		count      int
		maxRetries int
		retryDelay time.Duration
	}{
		{
			name:       "OpenAI API Integration",
			provider:   "openai/gpt-4o",
			envVar:     "OPENAI_API_KEY",
			topic:      "testing",
			count:      2,
			maxRetries: 3,
			retryDelay: time.Second * 2,
		},
		{
			name:       "Anthropic API Integration",
			provider:   "anthropic/claude-3-haiku-20240307",
			envVar:     "ANTHROPIC_API_KEY",
			topic:      "testing",
			count:      2,
			maxRetries: 3,
			retryDelay: time.Second * 2,
		},
		{
			name:       "Google API Integration",
			provider:   "google/gemini-1.5-flash",
			envVar:     "GOOGLE_API_KEY",
			topic:      "testing",
			count:      2,
			maxRetries: 3,
			retryDelay: time.Second * 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if API key not available
			if os.Getenv(tt.envVar) == "" {
				t.Skipf("Skipping %s: %s not set", tt.name, tt.envVar)
			}

			// Rate limiting: wait between tests
			time.Sleep(time.Second * 1)

			// Set the model for this test
			originalModel := os.Getenv("TIPS_MODEL")
			os.Setenv("TIPS_MODEL", tt.provider)
			defer os.Setenv("TIPS_MODEL", originalModel)

			// Test API call with retry logic
			var tips []Tip
			var err error

			for attempt := 0; attempt <= tt.maxRetries; attempt++ {
				if attempt > 0 {
					t.Logf("Retrying %s (attempt %d/%d)", tt.name, attempt, tt.maxRetries)
					time.Sleep(tt.retryDelay * time.Duration(attempt))
				}

				tipResponses, err := generateTips(tt.topic, tt.count)
				tips = convertToTips(tipResponses, tt.topic)
				if err == nil {
					break
				}

				// Log the error but continue retrying
				t.Logf("Attempt %d failed for %s: %v", attempt+1, tt.name, err)
			}

			// Final assertion
			if err != nil {
				t.Errorf("%s failed after %d retries: %v", tt.name, tt.maxRetries, err)
				return
			}

			// Validate response
			if len(tips) == 0 {
				t.Errorf("%s returned no tips", tt.name)
				return
			}

			if len(tips) != tt.count {
				t.Errorf("%s returned %d tips, expected %d", tt.name, len(tips), tt.count)
			}

			// Validate tip content
			for i, tip := range tips {
				if tip.Content == "" {
					t.Errorf("%s tip %d has empty content", tt.name, i)
				}
				if tip.Topic != tt.topic {
					t.Errorf("%s tip %d has wrong topic: got %s, want %s", tt.name, i, tip.Topic, tt.topic)
				}
				if tip.ID == "" {
					t.Errorf("%s tip %d has empty ID", tt.name, i)
				}
			}

			t.Logf("%s: Successfully generated %d tips", tt.name, len(tips))
		})
	}
}

// TestAPIErrorHandling tests how the application handles various API errors
func TestAPIErrorHandling(t *testing.T) {
	if !hasAnyAPIKey() {
		t.Skip("Skipping API error handling tests: no API keys found")
	}

	tests := []struct {
		name     string
		provider string
		envVar   string
		topic    string
		count    int
	}{
		{
			name:     "Large request handling",
			provider: getFirstAvailableProvider(),
			envVar:   getFirstAvailableAPIKey(),
			topic:    "very-specific-complex-advanced-topic-that-might-cause-issues",
			count:    1,
		},
		{
			name:     "Empty topic handling",
			provider: getFirstAvailableProvider(),
			envVar:   getFirstAvailableAPIKey(),
			topic:    "",
			count:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv(tt.envVar) == "" {
				t.Skipf("Skipping %s: %s not set", tt.name, tt.envVar)
			}

			time.Sleep(time.Second * 1) // Rate limiting

			tipResponses, err := generateTips(tt.topic, tt.count)
			tips := convertToTips(tipResponses, tt.topic)

			// For some error cases, we expect failure
			if tt.topic == "" {
				if err == nil {
					t.Logf("%s: Empty topic was handled gracefully", tt.name)
				} else {
					t.Logf("%s: Empty topic correctly returned error: %v", tt.name, err)
				}
			} else {
				if err != nil {
					t.Logf("%s: Expected request failed: %v", tt.name, err)
				} else {
					t.Logf("%s: Request succeeded with %d tips", tt.name, len(tips))
				}
			}
		})
	}
}

// Helper functions

func hasAnyAPIKey() bool {
	return os.Getenv("OPENAI_API_KEY") != "" ||
		os.Getenv("ANTHROPIC_API_KEY") != "" ||
		os.Getenv("GOOGLE_API_KEY") != ""
}

func getFirstAvailableProvider() string {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "openai"
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "anthropic"
	}
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return "google"
	}
	return "openai" // fallback
}

func getFirstAvailableAPIKey() string {
	if os.Getenv("OPENAI_API_KEY") != "" {
		return "OPENAI_API_KEY"
	}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return "ANTHROPIC_API_KEY"
	}
	if os.Getenv("GOOGLE_API_KEY") != "" {
		return "GOOGLE_API_KEY"
	}
	return "OPENAI_API_KEY" // fallback
}

// convertToTips converts TipResponse to Tip with proper topic assignment
func convertToTips(tipResponses []TipResponse, topic string) []Tip {
	tips := make([]Tip, len(tipResponses))
	for i, tr := range tipResponses {
		tips[i] = Tip{
			ID:        generateID(),
			Topic:     topic,
			Content:   tr.Content,
			CreatedAt: time.Now(),
		}
	}
	return tips
}

// generateID creates a simple unique ID for testing
func generateID() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}

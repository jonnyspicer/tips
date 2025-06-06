//go:build integration && api

package main

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRealAPIIntegration(t *testing.T) {
	if !hasAnyAPIKey() {
		t.Skip("Skipping API integration tests: no API keys found")
	}

	tests := []struct {
		name, provider, envVar, topic string
		count, maxRetries             int
		retryDelay                    time.Duration
	}{
		{"OpenAI API Integration", "openai/gpt-4o", "OPENAI_API_KEY", "testing", 2, 3, 2 * time.Second},
		{"Anthropic API Integration", "anthropic/claude-3-haiku-20240307", "ANTHROPIC_API_KEY", "testing", 2, 3, 2 * time.Second},
		{"Google API Integration", "google/gemini-1.5-flash", "GOOGLE_API_KEY", "testing", 2, 3, 2 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv(tt.envVar) == "" {
				t.Skipf("Skipping %s: %s not set", tt.name, tt.envVar)
			}
			time.Sleep(time.Second)
			originalModel := os.Getenv("TIPS_MODEL")
			os.Setenv("TIPS_MODEL", tt.provider)
			defer os.Setenv("TIPS_MODEL", originalModel)

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
				t.Logf("Attempt %d failed for %s: %v", attempt+1, tt.name, err)
			}

			if err != nil {
				t.Errorf("%s failed after %d retries: %v", tt.name, tt.maxRetries, err)
				return
			}
			if len(tips) == 0 {
				t.Errorf("%s returned no tips", tt.name)
				return
			}
			if len(tips) != tt.count {
				t.Errorf("%s returned %d tips, expected %d", tt.name, len(tips), tt.count)
			}
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

func TestAPIErrorHandling(t *testing.T) {
	if !hasAnyAPIKey() {
		t.Skip("Skipping API error handling tests: no API keys found")
	}

	tests := []struct {
		name, provider, envVar, topic string
		count                         int
	}{
		{"Large request handling", getFirstAvailableProvider(), getFirstAvailableAPIKey(), "very-specific-complex-advanced-topic-that-might-cause-issues", 1},
		{"Empty topic handling", getFirstAvailableProvider(), getFirstAvailableAPIKey(), "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if os.Getenv(tt.envVar) == "" {
				t.Skipf("Skipping %s: %s not set", tt.name, tt.envVar)
			}
			time.Sleep(time.Second)
			tipResponses, err := generateTips(tt.topic, tt.count)
			tips := convertToTips(tipResponses, tt.topic)
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
	return "openai"
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
	return "OPENAI_API_KEY"
}

func convertToTips(tipResponses []TipResponse, topic string) []Tip {
	tips := make([]Tip, len(tipResponses))
	for i, tr := range tipResponses {
		tips[i] = Tip{
			ID:        fmt.Sprintf("test-%d", time.Now().UnixNano()),
			Topic:     topic,
			Content:   tr.Content,
			CreatedAt: time.Now(),
		}
	}
	return tips
}
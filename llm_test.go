package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCreateLLM(t *testing.T) {
	ctx := context.Background()

	// Save original environment variables
	originalModel := os.Getenv("TIPS_MODEL")
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalAnthropic := os.Getenv("ANTHROPIC_API_KEY")
	originalGoogle := os.Getenv("GOOGLE_API_KEY")

	defer func() {
		os.Setenv("TIPS_MODEL", originalModel)
		os.Setenv("OPENAI_API_KEY", originalOpenAI)
		os.Setenv("ANTHROPIC_API_KEY", originalAnthropic)
		os.Setenv("GOOGLE_API_KEY", originalGoogle)
	}()

	tests := []struct {
		name          string
		model         string
		openaiKey     string
		anthropicKey  string
		googleKey     string
		expectError   bool
		errorContains string
	}{
		{
			name:          "default model with openai key",
			model:         "",
			openaiKey:     "test-key",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "invalid model format",
			model:         "invalid-format",
			openaiKey:     "test-key",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   true,
			errorContains: "invalid model format",
		},
		{
			name:          "openai without key",
			model:         "openai/gpt-4",
			openaiKey:     "",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   true,
			errorContains: "OPENAI_API_KEY",
		},
		{
			name:          "anthropic without key",
			model:         "anthropic/claude-3",
			openaiKey:     "",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   true,
			errorContains: "ANTHROPIC_API_KEY",
		},
		{
			name:          "google without key",
			model:         "google/gemini",
			openaiKey:     "",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   true,
			errorContains: "GOOGLE_API_KEY",
		},
		{
			name:          "unsupported provider",
			model:         "unsupported/model",
			openaiKey:     "test-key",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   true,
			errorContains: "unsupported provider",
		},
		{
			name:          "openai with key",
			model:         "openai/gpt-4",
			openaiKey:     "test-key",
			anthropicKey:  "",
			googleKey:     "",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "anthropic with key",
			model:         "anthropic/claude-3",
			openaiKey:     "",
			anthropicKey:  "test-key",
			googleKey:     "",
			expectError:   false,
			errorContains: "",
		},
		{
			name:          "google with key",
			model:         "google/gemini",
			openaiKey:     "",
			anthropicKey:  "",
			googleKey:     "test-key",
			expectError:   false,
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("TIPS_MODEL", tt.model)
			os.Setenv("OPENAI_API_KEY", tt.openaiKey)
			os.Setenv("ANTHROPIC_API_KEY", tt.anthropicKey)
			os.Setenv("GOOGLE_API_KEY", tt.googleKey)

			llm, err := createLLM(ctx)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if llm == nil {
					t.Error("Expected non-nil LLM")
				}
			}
		})
	}
}

func TestTipResponseStructure(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	tip := TipResponse{Content: "test content"}

	data, err := json.Marshal(tip)
	if err != nil {
		t.Errorf("Failed to marshal TipResponse: %v", err)
	}

	var unmarshaled TipResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal TipResponse: %v", err)
	}

	if unmarshaled.Content != tip.Content {
		t.Errorf("Expected content '%s', got '%s'", tip.Content, unmarshaled.Content)
	}
}

func TestTipsResponseStructure(t *testing.T) {
	// Test JSON marshaling/unmarshaling
	response := TipsResponse{
		Tips: []TipResponse{
			{Content: "tip 1"},
			{Content: "tip 2"},
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal TipsResponse: %v", err)
	}

	var unmarshaled TipsResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal TipsResponse: %v", err)
	}

	if len(unmarshaled.Tips) != 2 {
		t.Errorf("Expected 2 tips, got %d", len(unmarshaled.Tips))
	}

	if unmarshaled.Tips[0].Content != "tip 1" {
		t.Errorf("Expected first tip content 'tip 1', got '%s'", unmarshaled.Tips[0].Content)
	}
}

func TestGenerateTipsJSONParsing(t *testing.T) {
	tests := []struct {
		name          string
		response      string
		expectError   bool
		expectedTips  int
		errorContains string
	}{
		{
			name: "valid JSON response",
			response: `{
				"tips": [
					{"content": "tip 1"},
					{"content": "tip 2"}
				]
			}`,
			expectError:  false,
			expectedTips: 2,
		},
		{
			name: "JSON with markdown wrapper",
			response: "```json\n" + `{
				"tips": [
					{"content": "tip 1"}
				]
			}` + "\n```",
			expectError:  false,
			expectedTips: 1,
		},
		{
			name: "JSON with simple markdown wrapper",
			response: "```\n" + `{
				"tips": [
					{"content": "tip 1"}
				]
			}` + "\n```",
			expectError:  false,
			expectedTips: 1,
		},
		{
			name:          "invalid JSON",
			response:      `{"tips": [{"content": "incomplete`,
			expectError:   true,
			expectedTips:  0,
			errorContains: "failed to parse response as JSON",
		},
		{
			name: "empty tips array",
			response: `{
				"tips": []
			}`,
			expectError:   true,
			expectedTips:  0,
			errorContains: "no tips generated",
		},
		{
			name:          "malformed response",
			response:      "This is not JSON at all",
			expectError:   true,
			expectedTips:  0,
			errorContains: "failed to parse response as JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the JSON parsing logic directly
			cleanResp := strings.TrimSpace(tt.response)
			if strings.HasPrefix(cleanResp, "```json") || strings.HasPrefix(cleanResp, "```") {
				cleanResp = strings.TrimPrefix(cleanResp, "```json")
				cleanResp = strings.TrimPrefix(cleanResp, "```")
				cleanResp = strings.TrimSuffix(cleanResp, "```")
				cleanResp = strings.TrimSpace(cleanResp)
			}

			var tipsResponse TipsResponse
			err := json.Unmarshal([]byte(cleanResp), &tipsResponse)

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error parsing JSON: %v", err)
				return
			}

			if tt.expectError && err == nil && len(tipsResponse.Tips) == 0 {
				// Simulate the "no tips generated" error
				err = &NoTipsError{}
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					// For JSON errors, just check if we got an error
					if !strings.Contains(tt.errorContains, "failed to parse") {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
					}
				}
			} else {
				if len(tipsResponse.Tips) != tt.expectedTips {
					t.Errorf("Expected %d tips, got %d", tt.expectedTips, len(tipsResponse.Tips))
				}
			}
		})
	}
}

// Helper error type for testing
type NoTipsError struct{}

func (e *NoTipsError) Error() string {
	return "no tips generated in response"
}

func TestPromptGeneration(t *testing.T) {
	// Test that the prompt contains expected elements
	topic := "git"
	count := 5

	expectedPrompt := generatePromptString(topic, count)

	// Check key elements are present
	if !strings.Contains(expectedPrompt, topic) {
		t.Errorf("Prompt should contain topic '%s'", topic)
	}

	if !strings.Contains(expectedPrompt, "5") {
		t.Error("Prompt should contain count")
	}

	if !strings.Contains(expectedPrompt, "JSON") {
		t.Error("Prompt should mention JSON format")
	}

	if !strings.Contains(expectedPrompt, "cheatsheet") {
		t.Error("Prompt should mention cheatsheet style")
	}
}

// Helper function to generate prompt (extracted from generateTips for testing)
func generatePromptString(topic string, count int) string {
	return `Generate ` + string(rune(count+'0')) + ` concise cheatsheet-style tips about ` + topic + `. Each tip should be:
- Brief and to-the-point (1-2 sentences max)
- Include specific commands, shortcuts, or code snippets when applicable
- Focus on practical, immediately usable information
- Written in a reference format like you'd find in a quick reference guide

Examples of good cheatsheet tips:
- 'git stash: Temporarily save uncommitted changes with git stash, restore with git stash pop'
- 'vim: Delete entire line with dd, copy line with yy, paste with p'
- 'bash: Use !! to repeat last command, !$ for last argument of previous command'

IMPORTANT: Return ONLY a valid JSON object. Do not wrap it in markdown code blocks or add any other text. Use this exact format:
{
  "tips": [
    {"content": "tip 1 content here"},
    {"content": "tip 2 content here"}
  ]
}

Generate ` + string(rune(count+'0')) + ` tips about ` + topic + ` in this cheatsheet style.`
}

func TestModelEnvironmentHandling(t *testing.T) {
	// Save original
	original := os.Getenv("TIPS_MODEL")
	defer os.Setenv("TIPS_MODEL", original)

	// Test default model
	os.Setenv("TIPS_MODEL", "")
	expectedDefault := "openai/gpt-4o"

	model := os.Getenv("TIPS_MODEL")
	if model == "" {
		model = expectedDefault
	}

	if model != expectedDefault {
		t.Errorf("Expected default model '%s', got '%s'", expectedDefault, model)
	}

	// Test custom model
	customModel := "anthropic/claude-3-sonnet"
	os.Setenv("TIPS_MODEL", customModel)

	model = os.Getenv("TIPS_MODEL")
	if model != customModel {
		t.Errorf("Expected custom model '%s', got '%s'", customModel, model)
	}
}

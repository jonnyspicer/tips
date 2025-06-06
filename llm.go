package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

type TipResponse struct {
	Content string `json:"content"`
}

type TipsResponse struct {
	Tips []TipResponse `json:"tips"`
}

func createLLM(ctx context.Context) (llms.Model, error) {
	model := os.Getenv("TIPS_MODEL")
	if model == "" {
		model = "openai/gpt-4o"
	}

	parts := strings.Split(model, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid model format. Expected 'provider/model' (e.g., 'openai/gpt-4o')")
	}

	provider, modelName := parts[0], parts[1]

	switch provider {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set. Please set it with: export OPENAI_API_KEY='your-api-key'")
		}
		return openai.New(openai.WithModel(modelName), openai.WithToken(apiKey))
	case "anthropic":
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set. Please set it with: export ANTHROPIC_API_KEY='your-api-key'")
		}
		return anthropic.New(anthropic.WithModel(modelName), anthropic.WithToken(apiKey))
	case "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GOOGLE_API_KEY environment variable not set. Please set it with: export GOOGLE_API_KEY='your-api-key'")
		}
		return googleai.New(ctx, googleai.WithAPIKey(apiKey), googleai.WithDefaultModel(modelName))
	default:
		return nil, fmt.Errorf("unsupported provider: %s. Supported providers: openai, anthropic, google", provider)
	}
}

func generateTips(topic string, count int) ([]TipResponse, error) {
	ctx := context.Background()

	llm, err := createLLM(ctx)
	if err != nil {
		return nil, err
	}

	prompt := fmt.Sprintf(`Generate %d concise cheatsheet-style tips about %s. Each tip should be:
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

Generate %d tips about %s in this cheatsheet style.`, count, topic, count, topic)

	resp, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "API key") || strings.Contains(err.Error(), "authentication") {
			return nil, fmt.Errorf("invalid API key. Please check your API key environment variable")
		}
		if strings.Contains(err.Error(), "quota") || strings.Contains(err.Error(), "rate") {
			return nil, fmt.Errorf("API quota exceeded or rate limited. Please try again later")
		}
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	cleanResp := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanResp, "```json") || strings.HasPrefix(cleanResp, "```") {
		cleanResp = strings.TrimPrefix(cleanResp, "```json")
		cleanResp = strings.TrimPrefix(cleanResp, "```")
		cleanResp = strings.TrimSuffix(cleanResp, "```")
		cleanResp = strings.TrimSpace(cleanResp)
	}

	var tipsResponse TipsResponse
	if err := json.Unmarshal([]byte(cleanResp), &tipsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON. Raw response: %s. Cleaned response: %s. Error: %w", resp, cleanResp, err)
	}

	if len(tipsResponse.Tips) == 0 {
		return nil, fmt.Errorf("no tips generated in response")
	}

	return tipsResponse.Tips, nil
}

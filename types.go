package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Tip struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type TipsData struct {
	Tips []Tip `json:"tips"`
}

func getTipsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tips.json"), nil
}

func loadTips() (*TipsData, error) {
	filePath, err := getTipsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get tips file path: %w", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &TipsData{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tips file: %w", err)
	}

	if len(data) == 0 {
		return &TipsData{}, nil
	}

	var tipsData TipsData
	if err := json.Unmarshal(data, &tipsData); err != nil {
		return nil, fmt.Errorf("failed to parse tips file: %w", err)
	}

	return &tipsData, nil
}

func saveTips(tipsData *TipsData) error {
	if tipsData == nil {
		return fmt.Errorf("tips data cannot be nil")
	}

	filePath, err := getTipsFilePath()
	if err != nil {
		return fmt.Errorf("failed to get tips file path: %w", err)
	}

	data, err := json.MarshalIndent(tipsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tips data: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tips file: %w", err)
	}

	return nil
}

func (td *TipsData) addTip(topic, content string) {
	if topic == "" || content == "" {
		return
	}
	td.Tips = append(td.Tips, Tip{
		ID:        uuid.New().String(),
		Topic:     strings.TrimSpace(topic),
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now(),
	})
}

func (td *TipsData) removeTip(id string) bool {
	for i, tip := range td.Tips {
		if tip.ID == id {
			td.Tips = append(td.Tips[:i], td.Tips[i+1:]...)
			return true
		}
	}
	return false
}

func (td *TipsData) getRandomTip(topics []string) *Tip {
	if len(td.Tips) == 0 {
		return nil
	}

	filteredTips := td.Tips
	if len(topics) > 0 {
		topicMap := make(map[string]bool, len(topics))
		for _, topic := range topics {
			topicMap[topic] = true
		}

		filteredTips = nil
		for _, tip := range td.Tips {
			if topicMap[tip.Topic] {
				filteredTips = append(filteredTips, tip)
			}
		}
	}

	if len(filteredTips) == 0 {
		return nil
	}

	return &filteredTips[rand.Intn(len(filteredTips))]
}
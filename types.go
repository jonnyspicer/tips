// Package main provides data types and storage functions for the tips CLI tool.
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

// init initializes the random seed for tip selection
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Tip represents a single tip with metadata
type Tip struct {
	ID        string    `json:"id"`        // Unique identifier
	Topic     string    `json:"topic"`     // Topic category
	Content   string    `json:"content"`   // Tip content
	CreatedAt time.Time `json:"created_at"` // Creation timestamp
}

// TipsData holds the collection of all tips
type TipsData struct {
	Tips []Tip `json:"tips"` // Collection of tips
}

// getTipsFilePath returns the path to the tips storage file
func getTipsFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".tips.json"), nil
}

// loadTips loads tips from the storage file
func loadTips() (*TipsData, error) {
	filePath, err := getTipsFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get tips file path: %w", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &TipsData{Tips: []Tip{}}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tips file: %w", err)
	}

	if len(data) == 0 {
		return &TipsData{Tips: []Tip{}}, nil
	}

	var tipsData TipsData
	if err := json.Unmarshal(data, &tipsData); err != nil {
		return nil, fmt.Errorf("failed to parse tips file: %w", err)
	}

	return &tipsData, nil
}

// saveTips saves tips to the storage file
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

// addTip adds a new tip to the collection
func (td *TipsData) addTip(topic, content string) {
	if topic == "" || content == "" {
		return // Skip empty tips
	}
	tip := Tip{
		ID:        uuid.New().String(),
		Topic:     strings.TrimSpace(topic),
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now(),
	}
	td.Tips = append(td.Tips, tip)
}

// removeTip removes a tip by ID, returns true if found and removed
func (td *TipsData) removeTip(id string) bool {
	for i, tip := range td.Tips {
		if tip.ID == id {
			td.Tips = append(td.Tips[:i], td.Tips[i+1:]...)
			return true
		}
	}
	return false
}

// getRandomTip returns a random tip, optionally filtered by topics
func (td *TipsData) getRandomTip(topics []string) *Tip {
	if len(td.Tips) == 0 {
		return nil
	}

	var filteredTips []Tip
	if len(topics) == 0 {
		filteredTips = td.Tips
	} else {
		topicMap := make(map[string]bool)
		for _, topic := range topics {
			topicMap[topic] = true
		}

		for _, tip := range td.Tips {
			if topicMap[tip.Topic] {
				filteredTips = append(filteredTips, tip)
			}
		}
	}

	if len(filteredTips) == 0 {
		return nil
	}

	randomIndex := rand.Intn(len(filteredTips))
	return &filteredTips[randomIndex]
}
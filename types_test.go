package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTipsData_addTip(t *testing.T) {
	tests := []struct {
		name         string
		topic        string
		content      string
		expectAdded  bool
		expectedLen  int
	}{
		{
			name:        "valid tip",
			topic:       "git",
			content:     "Use git status to check repository status",
			expectAdded: true,
			expectedLen: 1,
		},
		{
			name:        "empty topic",
			topic:       "",
			content:     "Some content",
			expectAdded: false,
			expectedLen: 0,
		},
		{
			name:        "empty content",
			topic:       "git",
			content:     "",
			expectAdded: false,
			expectedLen: 0,
		},
		{
			name:        "whitespace topic",
			topic:       "  ",
			content:     "Some content",
			expectAdded: true, // addTip checks == "" before trimming, so "  " != ""
			expectedLen: 1,
		},
		{
			name:        "whitespace content",
			topic:       "git",
			content:     "  ",
			expectAdded: true, // addTip checks == "" before trimming, so "  " != ""
			expectedLen: 1,
		},
		{
			name:        "topic with whitespace",
			topic:       "  git  ",
			content:     "  Some content  ",
			expectAdded: true,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			td := &TipsData{}
			initialLen := len(td.Tips)
			
			td.addTip(tt.topic, tt.content)
			
			if len(td.Tips) != initialLen+tt.expectedLen {
				t.Errorf("Expected %d tips, got %d", initialLen+tt.expectedLen, len(td.Tips))
			}

			if tt.expectAdded && len(td.Tips) > 0 {
				tip := td.Tips[len(td.Tips)-1]
				
				// Check ID is valid UUID
				if _, err := uuid.Parse(tip.ID); err != nil {
					t.Errorf("Invalid UUID: %v", err)
				}
				
				// Check topic is trimmed
				if tip.Topic != strings.TrimSpace(tt.topic) {
					t.Errorf("Expected topic %q, got %q", strings.TrimSpace(tt.topic), tip.Topic)
				}
				
				// Check content is trimmed
				if tip.Content != strings.TrimSpace(tt.content) {
					t.Errorf("Expected content %q, got %q", strings.TrimSpace(tt.content), tip.Content)
				}
				
				// Check timestamp is recent
				if time.Since(tip.CreatedAt) > time.Second {
					t.Errorf("Timestamp too old: %v", tip.CreatedAt)
				}
			}
		})
	}
}

func TestTipsData_removeTip(t *testing.T) {
	td := &TipsData{}
	
	// Add some test tips
	tip1ID := "550e8400-e29b-41d4-a716-446655440001"
	tip2ID := "550e8400-e29b-41d4-a716-446655440002"
	
	td.Tips = []Tip{
		{ID: tip1ID, Topic: "git", Content: "tip 1", CreatedAt: time.Now()},
		{ID: tip2ID, Topic: "vim", Content: "tip 2", CreatedAt: time.Now()},
	}

	tests := []struct {
		name           string
		tipID          string
		expectedResult bool
		expectedLen    int
	}{
		{
			name:           "remove existing tip",
			tipID:          tip1ID,
			expectedResult: true,
			expectedLen:    1,
		},
		{
			name:           "remove non-existent tip",
			tipID:          "non-existent",
			expectedResult: false,
			expectedLen:    1,
		},
		{
			name:           "remove remaining tip",
			tipID:          tip2ID,
			expectedResult: true,
			expectedLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := td.removeTip(tt.tipID)
			
			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}
			
			if len(td.Tips) != tt.expectedLen {
				t.Errorf("Expected %d tips remaining, got %d", tt.expectedLen, len(td.Tips))
			}
			
			// Verify the tip was actually removed
			if tt.expectedResult {
				for _, tip := range td.Tips {
					if tip.ID == tt.tipID {
						t.Errorf("Tip %s was not removed", tt.tipID)
					}
				}
			}
		})
	}
}

func TestTipsData_getRandomTip(t *testing.T) {
	td := &TipsData{}
	
	// Test empty tips
	tip := td.getRandomTip(nil)
	if tip != nil {
		t.Error("Expected nil tip for empty tips data")
	}
	
	// Add test tips
	td.Tips = []Tip{
		{ID: "1", Topic: "git", Content: "git tip", CreatedAt: time.Now()},
		{ID: "2", Topic: "vim", Content: "vim tip", CreatedAt: time.Now()},
		{ID: "3", Topic: "git", Content: "another git tip", CreatedAt: time.Now()},
	}

	tests := []struct {
		name           string
		topics         []string
		expectedTopics []string
		shouldFind     bool
	}{
		{
			name:           "no filter",
			topics:         nil,
			expectedTopics: []string{"git", "vim"},
			shouldFind:     true,
		},
		{
			name:           "git filter",
			topics:         []string{"git"},
			expectedTopics: []string{"git"},
			shouldFind:     true,
		},
		{
			name:           "vim filter",
			topics:         []string{"vim"},
			expectedTopics: []string{"vim"},
			shouldFind:     true,
		},
		{
			name:           "multiple filters",
			topics:         []string{"git", "vim"},
			expectedTopics: []string{"git", "vim"},
			shouldFind:     true,
		},
		{
			name:           "non-existent topic",
			topics:         []string{"bash"},
			expectedTopics: []string{},
			shouldFind:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tip := td.getRandomTip(tt.topics)
			
			if !tt.shouldFind {
				if tip != nil {
					t.Error("Expected nil tip for non-matching filter")
				}
				return
			}
			
			if tip == nil {
				t.Error("Expected non-nil tip")
				return
			}
			
			// Check if returned tip matches expected topics
			topicFound := false
			for _, expectedTopic := range tt.expectedTopics {
				if tip.Topic == expectedTopic {
					topicFound = true
					break
				}
			}
			
			if len(tt.expectedTopics) > 0 && !topicFound {
				t.Errorf("Tip topic %q not in expected topics %v", tip.Topic, tt.expectedTopics)
			}
		})
	}
}

func TestGetTipsFilePath(t *testing.T) {
	path, err := getTipsFilePath()
	if err != nil {
		t.Fatalf("getTipsFilePath failed: %v", err)
	}
	
	if !strings.HasSuffix(path, ".tips.json") {
		t.Errorf("Expected path to end with .tips.json, got %s", path)
	}
	
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}
}

func TestLoadTips(t *testing.T) {
	// Test with non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		// Temporarily change home directory to avoid conflicts
		originalHome := os.Getenv("HOME")
		tmpDir := t.TempDir()
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)
		
		tipsData, err := loadTips()
		if err != nil {
			t.Fatalf("loadTips failed: %v", err)
		}
		
		if tipsData == nil {
			t.Fatal("Expected non-nil tips data")
		}
		
		if len(tipsData.Tips) != 0 {
			t.Errorf("Expected empty tips, got %d", len(tipsData.Tips))
		}
	})
	
	// Test with empty file
	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, ".tips.json")
		
		// Create empty file
		if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}
		
		// Mock getTipsFilePath
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)
		
		tipsData, err := loadTips()
		if err != nil {
			t.Fatalf("loadTips failed: %v", err)
		}
		
		if len(tipsData.Tips) != 0 {
			t.Errorf("Expected empty tips, got %d", len(tipsData.Tips))
		}
	})
	
	// Test with valid JSON
	t.Run("valid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, ".tips.json")
		
		// Create valid JSON file
		testData := TipsData{
			Tips: []Tip{
				{ID: "1", Topic: "test", Content: "test content", CreatedAt: time.Now()},
			},
		}
		
		jsonData, err := json.Marshal(testData)
		if err != nil {
			t.Fatalf("Failed to marshal test data: %v", err)
		}
		
		if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
		
		// Mock getTipsFilePath
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)
		
		tipsData, err := loadTips()
		if err != nil {
			t.Fatalf("loadTips failed: %v", err)
		}
		
		if len(tipsData.Tips) != 1 {
			t.Errorf("Expected 1 tip, got %d", len(tipsData.Tips))
		}
		
		if tipsData.Tips[0].Topic != "test" {
			t.Errorf("Expected topic 'test', got '%s'", tipsData.Tips[0].Topic)
		}
	})
}

func TestSaveTips(t *testing.T) {
	// Test with nil data
	t.Run("nil data", func(t *testing.T) {
		err := saveTips(nil)
		if err == nil {
			t.Error("Expected error for nil tips data")
		}
	})
	
	// Test with valid data
	t.Run("valid data", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Mock getTipsFilePath
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)
		
		testData := &TipsData{
			Tips: []Tip{
				{ID: "1", Topic: "test", Content: "test content", CreatedAt: time.Now()},
			},
		}
		
		err := saveTips(testData)
		if err != nil {
			t.Fatalf("saveTips failed: %v", err)
		}
		
		// Verify file was created
		filePath := filepath.Join(tmpDir, ".tips.json")
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Error("Tips file was not created")
		}
		
		// Verify content
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read saved file: %v", err)
		}
		
		var savedData TipsData
		if err := json.Unmarshal(data, &savedData); err != nil {
			t.Fatalf("Failed to unmarshal saved data: %v", err)
		}
		
		if len(savedData.Tips) != 1 {
			t.Errorf("Expected 1 tip in saved data, got %d", len(savedData.Tips))
		}
	})
}
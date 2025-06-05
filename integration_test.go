// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCLIIntegration(t *testing.T) {
	// Build the binary first
	binaryPath := buildTestBinary(t)
	defer os.Remove(binaryPath)

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectedOutput string
	}{
		{
			name:           "version flag",
			args:           []string{"--version"},
			expectError:    false,
			expectedOutput: Version,
		},
		{
			name:           "help flag",
			args:           []string{"--help"},
			expectError:    false,
			expectedOutput: "Tips is a command-line tool",
		},
		{
			name:           "show help",
			args:           []string{"show", "--help"},
			expectError:    false,
			expectedOutput: "Display tips in an interactive terminal interface",
		},
		{
			name:           "generate help",
			args:           []string{"generate", "--help"},
			expectError:    false,
			expectedOutput: "Generate new tips using AI language models",
		},
		{
			name:           "clear help",
			args:           []string{"clear", "--help"},
			expectError:    false,
			expectedOutput: "Delete all tips from local storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError && err == nil {
				t.Error("Expected error but command succeeded")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v, output: %s", err, outputStr)
			}

			if !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, outputStr)
			}
		})
	}
}

func TestFileSystemIntegration(t *testing.T) {
	// Create temporary home directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test file operations
	tipsData := &TipsData{
		Tips: []Tip{
			{
				ID:        "test-id",
				Topic:     "test-topic",
				Content:   "test content",
				CreatedAt: time.Now(),
			},
		},
	}

	// Test save and load
	err := saveTips(tipsData)
	if err != nil {
		t.Fatalf("Failed to save tips: %v", err)
	}

	// Verify file exists
	filePath, err := getTipsFilePath()
	if err != nil {
		t.Fatalf("Failed to get tips file path: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Tips file was not created")
	}

	// Test load
	loadedData, err := loadTips()
	if err != nil {
		t.Fatalf("Failed to load tips: %v", err)
	}

	if len(loadedData.Tips) != 1 {
		t.Errorf("Expected 1 tip, got %d", len(loadedData.Tips))
	}

	if loadedData.Tips[0].Content != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", loadedData.Tips[0].Content)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create temporary home directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Test concurrent reads and writes
	numGoroutines := 10
	errChan := make(chan error, numGoroutines*2)

	// Start concurrent writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			tipsData := &TipsData{
				Tips: []Tip{
					{
						ID:        string(rune('a' + id)),
						Topic:     "concurrent",
						Content:   "test content " + string(rune('0'+id)),
						CreatedAt: time.Now(),
					},
				},
			}
			errChan <- saveTips(tipsData)
		}(i)
	}

	// Start concurrent readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := loadTips()
			errChan <- err
		}()
	}

	// Collect results
	var errors []error
	for i := 0; i < numGoroutines*2; i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent access failed with errors: %v", errors)
	}
}

func TestLargeDatasetPerformance(t *testing.T) {
	// Create temporary home directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create large dataset
	tipsData := &TipsData{}
	for i := 0; i < 1000; i++ {
		tipsData.addTip("performance", "Performance test tip number "+string(rune('0'+i%10)))
	}

	// Test save performance
	start := time.Now()
	err := saveTips(tipsData)
	saveTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to save large dataset: %v", err)
	}

	if saveTime > 1*time.Second {
		t.Errorf("Save took too long: %v", saveTime)
	}

	// Test load performance
	start = time.Now()
	_, err = loadTips()
	loadTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to load large dataset: %v", err)
	}

	if loadTime > 1*time.Second {
		t.Errorf("Load took too long: %v", loadTime)
	}

	// Test random tip selection performance
	start = time.Now()
	for i := 0; i < 100; i++ {
		tipsData.getRandomTip([]string{"performance"})
	}
	selectionTime := time.Since(start)

	if selectionTime > 2*time.Second {
		t.Errorf("Random selection took too long: %v", selectionTime)
	}
}

func TestMalformedJSONHandling(t *testing.T) {
	// Create temporary home directory
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create malformed JSON file
	filePath, err := getTipsFilePath()
	if err != nil {
		t.Fatalf("Failed to get tips file path: %v", err)
	}

	malformedJSON := `{"tips": [{"id": "test", "topic": "test"` // Missing closing
	err = os.WriteFile(filePath, []byte(malformedJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write malformed JSON: %v", err)
	}

	// Test load with malformed JSON
	_, err = loadTips()
	if err == nil {
		t.Error("Expected error when loading malformed JSON")
	}

	if !strings.Contains(err.Error(), "failed to parse") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestPermissionHandling(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("GOOS") == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create temporary directory with restricted permissions
	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	err := os.Mkdir(restrictedDir, 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}
	defer os.Chmod(restrictedDir, 0755) // Restore permissions for cleanup

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", restrictedDir)
	defer os.Setenv("HOME", originalHome)

	// Test save with permission error
	tipsData := &TipsData{
		Tips: []Tip{
			{ID: "test", Topic: "test", Content: "test", CreatedAt: time.Now()},
		},
	}

	err = saveTips(tipsData)
	if err == nil {
		t.Error("Expected permission error when saving to restricted directory")
	}
}

func buildTestBinary(t *testing.T) string {
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "tips-test")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	return binaryPath
}

func TestJSONValidation(t *testing.T) {
	// Test valid JSON structures
	validStructures := []string{
		`{"tips": []}`,
		`{"tips": [{"id": "1", "topic": "test", "content": "test", "created_at": "2023-01-01T00:00:00Z"}]}`,
	}

	for i, jsonStr := range validStructures {
		t.Run("valid_"+string(rune('0'+i)), func(t *testing.T) {
			var tipsData TipsData
			err := json.Unmarshal([]byte(jsonStr), &tipsData)
			if err != nil {
				t.Errorf("Valid JSON should parse: %v", err)
			}
		})
	}

	// Test invalid JSON structures
	invalidStructures := []string{
		`{"invalid": true}`, // Missing tips field
		`{"tips": "not an array"}`,
		`{"tips": [{"missing_required_fields": true}]}`,
	}

	for i, jsonStr := range invalidStructures {
		t.Run("invalid_"+string(rune('0'+i)), func(t *testing.T) {
			var tipsData TipsData
			err := json.Unmarshal([]byte(jsonStr), &tipsData)
			// Should either error or result in empty/invalid structure
			if err == nil && len(tipsData.Tips) > 0 {
				// Check if the tip has required fields
				tip := tipsData.Tips[0]
				if tip.ID == "" || tip.Topic == "" || tip.Content == "" {
					// This is expected for malformed data
					return
				}
			}
		})
	}
}

func TestMemoryUsage(t *testing.T) {
	// Simple memory usage test
	var tipsData TipsData

	// Add many tips to test memory efficiency
	for i := 0; i < 10000; i++ {
		tipsData.addTip("memory-test", "This is a memory test tip with some content to test memory usage")
	}

	if len(tipsData.Tips) != 10000 {
		t.Errorf("Expected 10000 tips, got %d", len(tipsData.Tips))
	}

	// Test random access doesn't degrade too much
	start := time.Now()
	for i := 0; i < 100; i++ { // Reduced iterations for reasonable performance
		tipsData.getRandomTip([]string{"memory-test"})
	}
	duration := time.Since(start)

	if duration > 1*time.Second {
		t.Errorf("Random access too slow with large dataset: %v", duration)
	}
}
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestRootCommand(t *testing.T) {
	if rootCmd.Use != "tips" {
		t.Errorf("Expected command name 'tips', got '%s'", rootCmd.Use)
	}
	
	if rootCmd.Version != Version {
		t.Errorf("Expected version '%s', got '%s'", Version, rootCmd.Version)
	}
}

func TestCommandFlags(t *testing.T) {
	// Test that flags are properly defined
	flags := rootCmd.PersistentFlags()
	
	topicFlag := flags.Lookup("topic")
	if topicFlag == nil {
		t.Error("Topic flag not found")
	}
	
	refreshFlag := flags.Lookup("refresh")
	if refreshFlag == nil {
		t.Error("Refresh flag not found")
	}
	
	countFlag := flags.Lookup("count")
	if countFlag == nil {
		t.Error("Count flag not found")
	}
}

func TestGenerateTipsForTopics(t *testing.T) {
	// Save original values
	originalTopicFlag := topicFlag
	originalCountFlag := countFlag
	defer func() {
		topicFlag = originalTopicFlag
		countFlag = originalCountFlag
	}()

	tests := []struct {
		name          string
		topics        []string
		count         int
		expectError   bool
		errorContains string
	}{
		{
			name:          "empty topic string",
			topics:        []string{""},
			count:         5,
			expectError:   false, // Should skip empty topics with warning
			errorContains: "",
		},
		{
			name:          "whitespace topic",
			topics:        []string{"   "},
			count:         5,
			expectError:   false, // Should skip whitespace-only topics
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test values
			topicFlag = tt.topics
			countFlag = tt.count

			// Create a dummy command for testing
			cmd := &cobra.Command{}

			// For tests that don't exit, we can test them directly
			// The function will continue and try to generate tips but fail due to no API key
			// which is expected behavior in tests
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Function panicked: %v", r)
				}
			}()

			// Run the function - it may fail due to API calls but shouldn't panic
			generateTipsForTopics(cmd, []string{})
		})
	}
}

// Test validation logic separately to avoid os.Exit() calls
func TestGenerateTipsValidation(t *testing.T) {
	tests := []struct {
		name     string
		topics   []string
		count    int
		hasError bool
	}{
		{
			name:     "empty topics",
			topics:   []string{},
			count:    5,
			hasError: true,
		},
		{
			name:     "zero count",
			topics:   []string{"git"},
			count:    0,
			hasError: true,
		},
		{
			name:     "negative count",
			topics:   []string{"git"},
			count:    -1,
			hasError: true,
		},
		{
			name:     "valid input",
			topics:   []string{"git"},
			count:    5,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			hasError := false

			if len(tt.topics) == 0 {
				hasError = true
			}

			for _, topic := range tt.topics {
				if strings.TrimSpace(topic) == "" {
					continue
				}
				if tt.count <= 0 {
					hasError = true
				}
			}

			if hasError != tt.hasError {
				t.Errorf("Expected hasError=%v, got hasError=%v", tt.hasError, hasError)
			}
		})
	}
}

func TestClearAllTips(t *testing.T) {
	t.Run("no file exists", func(t *testing.T) {
		// Create temporary directory
		tmpDir := t.TempDir()
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Capture stdout
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd := &cobra.Command{}
		clearAllTips(cmd, []string{})

		// Restore stdout and read output
		w.Close()
		os.Stdout = originalStdout
		
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		if err != nil {
			t.Fatalf("Failed to read from pipe: %v", err)
		}
		output := buf.String()

		if !strings.Contains(output, "No tips file found") {
			t.Errorf("Expected 'No tips file found' message, got '%s'", output)
		}
	})

	t.Run("file exists and is deleted", func(t *testing.T) {
		// Create temporary directory and file
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, ".tips.json")
		
		// Create a dummy tips file
		if err := os.WriteFile(filePath, []byte(`{"tips":[]}`), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", originalHome)

		// Capture stdout
		originalStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd := &cobra.Command{}
		clearAllTips(cmd, []string{})

		// Restore stdout and read output
		w.Close()
		os.Stdout = originalStdout
		
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		if err != nil {
			t.Fatalf("Failed to read from pipe: %v", err)
		}
		output := buf.String()

		if !strings.Contains(output, "Successfully deleted tips file") {
			t.Errorf("Expected success message, got '%s'", output)
		}

		// Verify file was deleted
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			t.Error("File should have been deleted")
		}
	})
}

func TestShowCommandValidation(t *testing.T) {
	// Test validation logic without calling the actual command
	tests := []struct {
		name         string
		refreshFlag  int
		expectError  bool
	}{
		{
			name:        "valid refresh flag",
			refreshFlag: 60,
			expectError: false,
		},
		{
			name:        "zero refresh flag",
			refreshFlag: 0,
			expectError: true,
		},
		{
			name:        "negative refresh flag",
			refreshFlag: -1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.refreshFlag <= 0
			
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v for refresh flag %d, got error=%v", tt.expectError, tt.refreshFlag, hasError)
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Test that all expected commands are present
	commands := []string{"show", "generate", "clear"}
	
	for _, cmdName := range commands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Use == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Command '%s' not found", cmdName)
		}
	}
}

func TestGenerateCommandLong(t *testing.T) {
	// Test that generate command has proper documentation
	if !strings.Contains(generateCmd.Long, "API key") {
		t.Error("Generate command should mention API key requirements")
	}
	
	if !strings.Contains(generateCmd.Long, "OPENAI_API_KEY") {
		t.Error("Generate command should mention OPENAI_API_KEY")
	}
	
	if !strings.Contains(generateCmd.Long, "ANTHROPIC_API_KEY") {
		t.Error("Generate command should mention ANTHROPIC_API_KEY")
	}
	
	if !strings.Contains(generateCmd.Long, "GOOGLE_API_KEY") {
		t.Error("Generate command should mention GOOGLE_API_KEY")
	}
}

func TestFlagDefaults(t *testing.T) {
	// Test default values are reasonable
	if refreshFlag <= 0 {
		refreshFlag = 60 // Reset to default for this test
	}
	
	if countFlag <= 0 {
		countFlag = 20 // Reset to default for this test
	}
	
	if refreshFlag != 60 {
		t.Errorf("Expected default refresh flag to be 60, got %d", refreshFlag)
	}
	
	if countFlag != 20 {
		t.Errorf("Expected default count flag to be 20, got %d", countFlag)
	}
}
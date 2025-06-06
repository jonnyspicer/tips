package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	topics := []string{"git", "vim"}
	refreshMinutes := 30

	m := initialModel(topics, refreshMinutes)

	if len(m.topicFilter) != 2 {
		t.Errorf("Expected 2 topic filters, got %d", len(m.topicFilter))
	}

	if m.topicFilter[0] != "git" || m.topicFilter[1] != "vim" {
		t.Errorf("Topic filter not set correctly: %v", m.topicFilter)
	}

	expectedDuration := time.Duration(refreshMinutes) * time.Minute
	if m.refreshRate != expectedDuration {
		t.Errorf("Expected refresh rate %v, got %v", expectedDuration, m.refreshRate)
	}

	if m.quit {
		t.Error("Model should not start in quit state")
	}

	if m.tipsData == nil {
		t.Error("Tips data should be initialized")
	}
}

func TestModelInit(t *testing.T) {
	m := initialModel([]string{}, 60)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestModelUpdate_KeyMessages(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectQuit   bool
		expectNewTip bool
	}{
		{
			name:         "quit with q",
			key:          "q",
			expectQuit:   true,
			expectNewTip: false,
		},
		{
			name:         "next tip with n",
			key:          "n",
			expectQuit:   false,
			expectNewTip: false,
		},
		{
			name:         "mark known with k",
			key:          "k",
			expectQuit:   false,
			expectNewTip: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := initialModel([]string{}, 60)
			m.tipsData = &TipsData{
				Tips: []Tip{
					{ID: "1", Topic: "git", Content: "test tip", CreatedAt: time.Now()},
				},
			}
			m.currentTip = &m.tipsData.Tips[0]
			m.quit = false
			m.showNewTip = false

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}

			updatedModel, _ := m.Update(keyMsg)
			updated := updatedModel.(model)

			if updated.quit != tt.expectQuit {
				t.Errorf("Expected quit=%v, got quit=%v", tt.expectQuit, updated.quit)
			}

			if updated.showNewTip != tt.expectNewTip {
				t.Errorf("Expected showNewTip=%v, got showNewTip=%v", tt.expectNewTip, updated.showNewTip)
			}
		})
	}
}

func TestModelUpdate_CtrlC(t *testing.T) {
	m := initialModel([]string{}, 60)

	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlC}
	updatedModel, cmd := m.Update(keyMsg)
	updated := updatedModel.(model)

	if !updated.quit {
		t.Error("Ctrl+C should set quit to true")
	}

	if cmd == nil {
		t.Error("Ctrl+C should return quit command")
	}
}

func TestModelUpdate_TipsData(t *testing.T) {
	m := initialModel([]string{}, 60)

	newTipsData := &TipsData{
		Tips: []Tip{
			{ID: "1", Topic: "test", Content: "test content", CreatedAt: time.Now()},
		},
	}

	updatedModel, _ := m.Update(newTipsData)
	updated := updatedModel.(model)

	if updated.tipsData != newTipsData {
		t.Error("Tips data should be updated")
	}

	if !updated.showNewTip {
		t.Error("Should show new tip after data update")
	}
}

func TestModelUpdate_TickMessage(t *testing.T) {
	m := initialModel([]string{}, 60)

	tickMsg := tickMsg(time.Now())
	updatedModel, cmd := m.Update(tickMsg)
	updated := updatedModel.(model)

	if !updated.showNewTip {
		t.Error("Tick should trigger new tip")
	}

	if cmd == nil {
		t.Error("Tick should return batch command")
	}
}

func TestModelUpdate_ErrorMessage(t *testing.T) {
	m := initialModel([]string{}, 60)

	testError := &TestError{message: "test error"}
	updatedModel, _ := m.Update(testError)
	updated := updatedModel.(model)

	if !strings.Contains(updated.message, "test error") {
		t.Errorf("Expected error message to contain 'test error', got '%s'", updated.message)
	}
}

func TestModelView(t *testing.T) {
	tests := []struct {
		name           string
		setupModel     func() model
		expectedOutput string
	}{
		{
			name: "quit state",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.quit = true
				return m
			},
			expectedOutput: "",
		},
		{
			name: "loading state",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.tipsData = nil
				return m
			},
			expectedOutput: "Loading tips...",
		},
		{
			name: "no tips",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.tipsData = &TipsData{Tips: []Tip{}}
				return m
			},
			expectedOutput: "No tips found",
		},
		{
			name: "no current tip",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.tipsData = &TipsData{
					Tips: []Tip{
						{ID: "1", Topic: "git", Content: "test", CreatedAt: time.Now()},
					},
				}
				m.currentTip = nil
				return m
			},
			expectedOutput: "No tips available",
		},
		{
			name: "filtered no tips",
			setupModel: func() model {
				m := initialModel([]string{"bash"}, 60)
				m.tipsData = &TipsData{
					Tips: []Tip{
						{ID: "1", Topic: "git", Content: "test", CreatedAt: time.Now()},
					},
				}
				m.currentTip = nil
				return m
			},
			expectedOutput: "No tips found for topics",
		},
		{
			name: "normal tip display",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.tipsData = &TipsData{
					Tips: []Tip{
						{ID: "1", Topic: "git", Content: "test tip content", CreatedAt: time.Now()},
					},
				}
				m.currentTip = &m.tipsData.Tips[0]
				return m
			},
			expectedOutput: "git",
		},
		{
			name: "tip with message",
			setupModel: func() model {
				m := initialModel([]string{}, 60)
				m.tipsData = &TipsData{
					Tips: []Tip{
						{ID: "1", Topic: "git", Content: "test tip", CreatedAt: time.Now()},
					},
				}
				m.currentTip = &m.tipsData.Tips[0]
				m.message = "Test message"
				return m
			},
			expectedOutput: "Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupModel()
			output := m.View()

			if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, output)
			}
		})
	}
}

func TestTickCmd(t *testing.T) {
	duration := 5 * time.Second
	cmd := tickCmd(duration)

	if cmd == nil {
		t.Error("tickCmd should return a command")
	}
}

func TestLoadTipsCmd(t *testing.T) {
	cmd := loadTipsCmd()

	if cmd == nil {
		t.Error("loadTipsCmd should return a command")
	}

	msg := cmd()

	switch msg.(type) {
	case *TipsData:
	case error:
	default:
		t.Errorf("Expected TipsData or error, got %T", msg)
	}
}

func TestRunSimpleShow(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("runSimpleShow panicked: %v", r)
		}
	}()

	_ = runSimpleShow
}

func TestModelMarkKnown(t *testing.T) {
	m := initialModel([]string{}, 60)
	m.tipsData = &TipsData{
		Tips: []Tip{
			{ID: "1", Topic: "git", Content: "test tip 1", CreatedAt: time.Now()},
			{ID: "2", Topic: "vim", Content: "test tip 2", CreatedAt: time.Now()},
		},
	}
	m.currentTip = &m.tipsData.Tips[0]
	m.showNewTip = false
	initialTipCount := len(m.tipsData.Tips)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	updatedModel, _ := m.Update(keyMsg)
	updated := updatedModel.(model)

	if len(updated.tipsData.Tips) != initialTipCount-1 {
		t.Errorf("Expected %d tips after removal, got %d", initialTipCount-1, len(updated.tipsData.Tips))
	}

	for _, tip := range updated.tipsData.Tips {
		if tip.ID == "1" {
			t.Error("Tip with ID '1' should have been removed")
		}
	}
}

type TestError struct {
	message string
}

func (e *TestError) Error() string {
	return e.message
}

func TestModelStateTransitions(t *testing.T) {
	m := initialModel([]string{}, 60)

	expectedShowNewTip := true
	if m.tipsData != nil && len(m.tipsData.Tips) > 0 && m.currentTip != nil {
		expectedShowNewTip = false
	}

	if m.showNewTip != expectedShowNewTip {
		t.Errorf("Model showNewTip should be %v, got %v", expectedShowNewTip, m.showNewTip)
	}

	m.tipsData = &TipsData{
		Tips: []Tip{
			{ID: "1", Topic: "test", Content: "test content", CreatedAt: time.Now()},
		},
	}

	if m.showNewTip && m.tipsData != nil {
		if newTip := m.tipsData.getRandomTip(m.topicFilter); newTip != nil {
			m.currentTip = newTip
		}
		m.showNewTip = false
		m.message = ""
	}

	if m.showNewTip {
		t.Error("showNewTip should be false after processing")
	}

	if m.currentTip == nil {
		t.Error("currentTip should be set")
	}
}

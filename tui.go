// Package main provides the terminal user interface for the tips CLI tool.
package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// tickMsg represents a timer tick for automatic refresh
type tickMsg time.Time

// model represents the application state for the TUI
type model struct {
	tipsData    *TipsData     // All loaded tips
	currentTip  *Tip          // Currently displayed tip
	topicFilter []string      // Topics to filter by
	refreshRate time.Duration // How often to refresh tips
	lastRefresh time.Time     // Last refresh timestamp
	quit        bool          // Whether to quit the application
	showNewTip  bool          // Whether to select a new tip
	message     string        // Status message to display
}

// initialModel creates the initial application state
func initialModel(topics []string, refreshMinutes int) model {
	// Force color support
	lipgloss.SetColorProfile(termenv.ANSI256)
	
	// Load tips synchronously at startup to avoid race conditions
	tipsData, err := loadTips()
	if err != nil {
		fmt.Printf("Error loading tips: %v\n", err)
		tipsData = &TipsData{Tips: []Tip{}}
	}
	
	return model{
		topicFilter: topics,
		refreshRate: time.Duration(refreshMinutes) * time.Minute,
		showNewTip:  true,
		tipsData:    tipsData,
	}
}

// Init initializes the model (required by BubbleTea)
func (m model) Init() tea.Cmd {
	return tickCmd(m.refreshRate)
}

// tickCmd creates a command that sends a tick message after duration d
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadTipsCmd creates a command to load tips from storage
func loadTipsCmd() tea.Cmd {
	return func() tea.Msg {
		tipsData, err := loadTips()
		if err != nil {
			return err
		}
		return tipsData
	}
}

// Update handles messages and updates the model (required by BubbleTea)
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quit = true
			return m, tea.Quit
		}
		
		switch msg.String() {
		case "q":
			m.quit = true
			return m, tea.Quit
		case "n":
			m.showNewTip = true
		case "k":
			if m.currentTip != nil && m.tipsData != nil {
				if m.tipsData.removeTip(m.currentTip.ID) {
					if err := saveTips(m.tipsData); err != nil {
						m.message = fmt.Sprintf("Error saving: %v", err)
					} else {
						m.message = "Tip marked as known!"
						m.showNewTip = true
					}
				}
			}
		}

	case *TipsData:
		m.tipsData = msg
		m.showNewTip = true
		return m, nil

	case tickMsg:
		m.showNewTip = true
		m.lastRefresh = time.Time(msg)
		return m, tea.Batch(
			tickCmd(m.refreshRate),
			loadTipsCmd(),
		)

	case error:
		m.message = fmt.Sprintf("Error: %v", msg)
		return m, nil
	}

	if m.showNewTip && m.tipsData != nil {
		newTip := m.tipsData.getRandomTip(m.topicFilter)
		if newTip != nil {
			m.currentTip = newTip
		}
		m.showNewTip = false
		m.message = ""
	}

	return m, nil
}

// View renders the current state (required by BubbleTea)
func (m model) View() string {
	if m.quit {
		return ""
	}

	if m.tipsData == nil {
		return "Loading tips..."
	}

	if len(m.tipsData.Tips) == 0 {
		return "No tips found. Generate some tips first using: tips generate -t <topic>\n\nPress 'q' to quit."
	}

	if m.currentTip == nil {
		if len(m.topicFilter) > 0 {
			return fmt.Sprintf("No tips found for topics: %v\n\nPress 'q' to quit.", m.topicFilter)
		}
		return "No tips available!\n\nPress 'q' to quit."
	}

	// Styles with explicit color codes
	topicStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).  // Cyan
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))   // White

	controlsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#808080")).  // Gray
		MarginTop(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).  // Yellow
		MarginTop(1)

	// Build the view
	var output string
	output += topicStyle.Render(fmt.Sprintf("[%s]", m.currentTip.Topic)) + " "
	output += contentStyle.Render(m.currentTip.Content) + "\n"
	
	refreshMinutes := int(m.refreshRate.Minutes())
	output += controlsStyle.Render(fmt.Sprintf("n:next | k:known | q:quit | refresh:%dm", refreshMinutes))

	if m.message != "" {
		output += "\n" + messageStyle.Render(m.message)
	}

	return output
}

// getCurrentTipID returns a short ID for the current tip (for debugging)
func getCurrentTipID(tip *Tip) string {
	if tip == nil {
		return "nil"
	}
	return tip.ID[:8] // First 8 chars of UUID
}

// runBubbleTeaShow starts the interactive TUI mode
func runBubbleTeaShow() error {
	// Try to create the program
	m := initialModel(topicFlag, refreshFlag)
	p := tea.NewProgram(m, tea.WithInput(os.Stdin))
	_, err := p.Run()
	
	// If TTY error, fall back to simple mode
	if err != nil {
		return runSimpleShow()
	}
	
	return err
}

// runSimpleShow provides a fallback non-interactive mode
func runSimpleShow() error {
	tipsData, err := loadTips()
	if err != nil {
		return fmt.Errorf("error loading tips: %w", err)
	}

	if len(tipsData.Tips) == 0 {
		fmt.Println("No tips found. Generate some tips first using: tips generate -t <topic>")
		return nil
	}

	tip := tipsData.getRandomTip(topicFlag)
	if tip == nil {
		if len(topicFlag) > 0 {
			fmt.Printf("No tips found for topics: %v\n", topicFlag)
		} else {
			fmt.Println("No tips available!")
		}
		return nil
	}

	// Use lipgloss for styling even in simple mode
	topicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))

	fmt.Printf("%s %s\n", topicStyle.Render(fmt.Sprintf("[%s]", tip.Topic)), tip.Content)
	fmt.Printf("\n%s\n", controlsStyle.Render(fmt.Sprintf("n:next | k:known | q:quit | refresh:%dm", refreshFlag)))
	return nil
}
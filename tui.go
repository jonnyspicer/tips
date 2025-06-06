package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type tickMsg time.Time

type model struct {
	tipsData    *TipsData
	currentTip  *Tip
	topicFilter []string
	refreshRate time.Duration
	lastRefresh time.Time
	quit        bool
	showNewTip  bool
	message     string
}

func initialModel(topics []string, refreshMinutes int) model {
	lipgloss.SetColorProfile(termenv.ANSI256)

	tipsData, err := loadTips()
	if err != nil {
		fmt.Printf("Error loading tips: %v\n", err)
		tipsData = &TipsData{}
	}

	m := model{
		topicFilter: topics,
		refreshRate: time.Duration(refreshMinutes) * time.Minute,
		showNewTip:  true,
		tipsData:    tipsData,
	}

	if tipsData != nil && len(tipsData.Tips) > 0 {
		if tip := tipsData.getRandomTip(topics); tip != nil {
			m.currentTip = tip
			m.showNewTip = false
		}
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tickCmd(m.refreshRate)
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func loadTipsCmd() tea.Cmd {
	return func() tea.Msg {
		tipsData, err := loadTips()
		if err != nil {
			return err
		}
		return tipsData
	}
}

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
		return m, tea.Batch(tickCmd(m.refreshRate), loadTipsCmd())

	case error:
		m.message = fmt.Sprintf("Error: %v", msg)
		return m, nil
	}

	if m.showNewTip && m.tipsData != nil {
		if newTip := m.tipsData.getRandomTip(m.topicFilter); newTip != nil {
			m.currentTip = newTip
		}
		m.showNewTip = false
		m.message = ""
	}

	return m, nil
}

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

	topicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).MarginTop(1)
	messageStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).MarginTop(1)

	output := topicStyle.Render(fmt.Sprintf("[%s]", m.currentTip.Topic)) + " " +
		contentStyle.Render(m.currentTip.Content) + "\n" +
		controlsStyle.Render(fmt.Sprintf("n:next | k:known | q:quit | refresh:%dm", int(m.refreshRate.Minutes())))

	if m.message != "" {
		output += "\n" + messageStyle.Render(m.message)
	}

	return output
}

func runBubbleTeaShow() error {
	m := initialModel(topicFlag, refreshFlag)
	p := tea.NewProgram(m, tea.WithInput(os.Stdin))
	_, err := p.Run()

	if err != nil {
		return runSimpleShow()
	}

	return err
}

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

	topicStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("cyan")).Bold(true)
	controlsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("gray"))

	fmt.Printf("%s %s\n", topicStyle.Render(fmt.Sprintf("[%s]", tip.Topic)), tip.Content)
	fmt.Printf("\n%s\n", controlsStyle.Render(fmt.Sprintf("n:next | k:known | q:quit | refresh:%dm", refreshFlag)))
	return nil
}

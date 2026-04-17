package glagentgui

import (
	"fmt"
	"os"

	"glagent/src/modules/agentMod"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func StartGUI() {
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func InitialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message... (Alt+Enter for new line)"
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New()
	vp.SetWidth(30)
	vp.SetHeight(5)
	vp.SetContent(`Welcome to GlAgent! Type your prompt below.`)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		input:    ta,
		viewport: vp,
		spinner:  s,
		messages: []message{},
		chat:     agentMod.NewChatSession(10),
	}
}

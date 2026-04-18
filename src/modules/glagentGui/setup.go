package glagentgui

import (
	"glagent/src/modules/agentMod"
	"glagent/src/modules/computer"
	"glagent/src/modules/sessionstore"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type StartOptions struct {
	ContinueSessionID string
	SessionID         string
}

func StartGUI(options StartOptions) error {
	initialModel, err := InitialModel(options)
	if err != nil {
		return err
	}

	p := tea.NewProgram(initialModel)
	_, err = p.Run()
	return err
}

func InitialModel(options StartOptions) (model, error) {
	if options.ContinueSessionID != "" {
		stored, err := sessionstore.Load(options.ContinueSessionID)
		if err != nil {
			return model{}, err
		}
		return modelFromStoredSession(stored), nil
	}

	sessionID := options.SessionID
	if sessionID == "" {
		sessionID = sessionstore.NewID()
	}

	ta := textarea.New()
	ta.Placeholder = "Send a message... (/ for commands, Alt+Enter for new line)"
	ta.Focus()
	ta.Prompt = "| "
	ta.CharLimit = 4096
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New()
	vp.SetWidth(30)
	vp.SetHeight(5)
	vp.SetContent(`Welcome to GlAgent! Type your prompt below.
Type / to see available commands, or /help for details.
Workspace command execution is enabled by default.
Use /computer full only if you want broader shell control.`)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(orangeBright)

	m := model{
		input:          ta,
		viewport:       vp,
		spinner:        s,
		messages:       []message{},
		selectorItems:  []selectorItem{},
		chat:           agentMod.NewChatSession(10),
		sessionID:      sessionID,
		permissionMode: computer.PermissionWorkspace,
	}
	m.addSystemMessage("Session started: " + sessionID)
	m.addSystemMessage("Computer control is in workspace mode. Use /computer off, /computer workspace, or /computer full.")
	_ = m.saveSession()

	return m, nil
}

package glagentgui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"glagent/src/modules/agentMod"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
)

type focusArea int

const (
	focusInput focusArea = iota
	focusProvider
	focusMode
)

type message struct {
	Role    string
	Content string
	Time    time.Time
}

type aiResponseMsg struct {
	Text string
	Err  error
}

type model struct {
	width  int
	height int

	viewport viewport.Model
	input    textarea.Model
	help     help.Model
	spinner  spinner.Model

	messages []message

	providers        []string
	selectedProvider int

	modes        []string
	selectedMode int

	focus focusArea

	showProviderPanel bool
	showModePanel     bool
	loading           bool

	chat *agentMod.ChatSession
	err  error
}

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			MarginBottom(1)

	botMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			MarginBottom(1)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	sysMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			MarginBottom(1)
)

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.viewport.SetWidth(msg.Width - 4)
		m.viewport.SetHeight(msg.Height - 12)
		m.input.SetWidth(msg.Width - 4)

		m.updateViewport()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "alt+enter", "shift+enter":
			m.input.InsertString("\n")
		case "enter":
			v := strings.TrimSpace(m.input.Value())
			if v != "" && !m.loading {
				m.input.Reset()
				
				if strings.HasPrefix(v, "/") {
					m.handleSlashCommand(v)
					return m, nil
				}

				m.messages = append(m.messages, message{
					Role:    "User",
					Content: v,
					Time:    time.Now(),
				})
				m.input.Reset()
				m.loading = true
				m.chat.AddUserMessage(v)
				m.updateViewport()
				return m, tea.Batch(
					m.spinner.Tick,
					m.makeAgentCall(v),
				)
			}
		}

	case aiResponseMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.chat.AddAssistantMessage(msg.Text)
			m.messages = append(m.messages, message{
				Role:    "Assistant",
				Content: msg.Text,
				Time:    time.Now(),
			})
		}
		m.updateViewport()
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.loading {
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleSlashCommand(cmdStr string) {
	parts := strings.SplitN(cmdStr, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "/agent":
		if arg == "" {
			m.messages = append(m.messages, message{Role: "System", Content: "Usage: /agent <provider> (e.g. gemini, openai, anthropic, groq, ollama)", Time: time.Now()})
		} else {
			os.Setenv("AI_PROVIDER", arg)
			m.messages = append(m.messages, message{Role: "System", Content: fmt.Sprintf("AI Provider set to '%s'", arg), Time: time.Now()})
		}
	case "/mode":
		if arg == "" {
			m.messages = append(m.messages, message{Role: "System", Content: "Usage: /mode <model> (e.g. gemini-2.5-flash)", Time: time.Now()})
		} else {
			os.Setenv("AI_MODEL", arg)
			m.messages = append(m.messages, message{Role: "System", Content: fmt.Sprintf("AI Model set to '%s'", arg), Time: time.Now()})
		}
	case "/help":
		helpText := "Available commands:\n/agent <provider> - Switch AI provider\n/mode <model> - Switch AI model\n/help - Show this help message"
		m.messages = append(m.messages, message{Role: "System", Content: helpText, Time: time.Now()})
	default:
		m.messages = append(m.messages, message{Role: "System", Content: fmt.Sprintf("Unknown command: %s. Type /help for a list of commands.", cmd), Time: time.Now()})
	}
	m.updateViewport()
}

func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("Loading...")
		v.AltScreen = true
		return v
	}

	title := titleStyle.Render(" GlAgent ")
	header := fmt.Sprintf("%s\n\n", title)

	vp := viewportStyle.Render(m.viewport.View())
	
	status := " Ready"
	if m.loading {
		status = fmt.Sprintf(" %s Thinking...", m.spinner.View())
	}

	if m.err != nil {
		status = errorStyle.Render(fmt.Sprintf(" Error: %v", m.err))
	}

	in := inputStyle.Render(m.input.View())

	content := lipgloss.JoinVertical(lipgloss.Left, header, vp, status, in)
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m *model) updateViewport() {
	var b strings.Builder
	for _, msg := range m.messages {
		if msg.Role == "User" {
			b.WriteString(userMsgStyle.Render("You: "))
			b.WriteString(msg.Content)
		} else if msg.Role == "System" {
			b.WriteString(sysMsgStyle.Render(fmt.Sprintf("• %s", msg.Content)))
		} else {
			b.WriteString(botMsgStyle.Render("GlAgent: "))
			
			// Try to render markdown
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(m.viewport.Width()),
			)
			if err == nil {
				out, err := renderer.Render(msg.Content)
				if err == nil {
					b.WriteString("\n" + strings.TrimSpace(out))
				} else {
					b.WriteString(msg.Content)
				}
			} else {
				b.WriteString(msg.Content)
			}
		}
		b.WriteString("\n\n")
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m model) makeAgentCall(prompt string) tea.Cmd {
	return func() tea.Msg {
		resp, err := agentMod.AskAI(prompt)
		return aiResponseMsg{Text: resp, Err: err}
	}
}
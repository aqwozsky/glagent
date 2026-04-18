package glagentgui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"glagent/src/modules/agentMod"
	"glagent/src/modules/computer"
	"glagent/src/modules/memory"
	"glagent/src/modules/sessionstore"
	"glagent/src/prompts"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/glamour"
	"github.com/joho/godotenv"
)

type focusArea int

const (
	focusInput focusArea = iota
	focusProvider
	focusMode
)

const maxAgentSteps = 4

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

	showSelector      bool
	selectorItems     []selectorItem
	selectorCursor    int
	selectorModeState selectorMode

	chat           *agentMod.ChatSession
	err            error
	sessionID      string
	permissionMode computer.PermissionMode
}

var (
	violetMid    = lipgloss.Color("#6C3BAA")
	violetLight  = lipgloss.Color("#B07CD8")
	violetPale   = lipgloss.Color("#D4BEE4")
	orangeBright = lipgloss.Color("#FF8C42")
	orangeWarm   = lipgloss.Color("#FFB347")
	textLight    = lipgloss.Color("#FFFDF5")
	textMuted    = lipgloss.Color("#8B7DA8")
	errorRed     = lipgloss.Color("#FF4757")
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(textLight).
			Background(violetMid).
			Padding(0, 2).
			Bold(true)

	userMsgStyle = lipgloss.NewStyle().
			Foreground(orangeBright).
			Bold(true).
			MarginBottom(1)

	botMsgStyle = lipgloss.NewStyle().
			Foreground(violetLight).
			Bold(true).
			MarginBottom(1)

	viewportStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(violetMid)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(violetLight)

	inputActiveStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(orangeBright)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorRed).
			Bold(true)

	sysMsgStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			Italic(true).
			MarginBottom(1)

	statusReadyStyle = lipgloss.NewStyle().
				Foreground(violetPale)

	statusLoadingStyle = lipgloss.NewStyle().
				Foreground(orangeWarm).
				Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(textMuted).
			PaddingLeft(1)
)

func modelFromStoredSession(stored *sessionstore.Session) model {
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

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(orangeBright)

	messages := make([]message, 0, len(stored.Messages))
	for _, msg := range stored.Messages {
		messages = append(messages, message{
			Role:    msg.Role,
			Content: msg.Content,
			Time:    msg.Time,
		})
	}

	permissionMode := computer.ParsePermissionMode(stored.PermissionMode)
	if permissionMode == "" {
		permissionMode = computer.PermissionWorkspace
	}

	return model{
		input:          ta,
		viewport:       vp,
		spinner:        s,
		messages:       messages,
		selectorItems:  []selectorItem{},
		chat:           agentMod.NewChatSessionFromEntries(stored.ChatEntries, 10),
		sessionID:      stored.ID,
		permissionMode: permissionMode,
	}
}

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
		m.input.SetWidth(msg.Width - 4)
		m.recalcViewportHeight()
		m.updateViewport()

	case tea.KeyPressMsg:
		if m.showSelector {
			switch msg.String() {
			case "up":
				if m.selectorCursor > 0 {
					m.selectorCursor--
				}
				return m, nil
			case "down":
				if m.selectorCursor < len(m.selectorItems)-1 {
					m.selectorCursor++
				}
				return m, nil
			case "tab":
				m.completeSelection()
				return m, nil
			case "esc":
				m.showSelector = false
				m.selectorCursor = 0
				m.recalcViewportHeight()
				return m, nil
			}
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.showSelector {
				m.showSelector = false
				m.selectorCursor = 0
				return m, nil
			}
			return m, tea.Quit
		case "alt+enter", "shift+enter":
			m.input.InsertString("\n")
		case "enter":
			if m.showSelector && len(m.selectorItems) > 0 {
				m.completeSelection()
				return m, nil
			}

			v := strings.TrimSpace(m.input.Value())
			if v != "" && !m.loading {
				m.input.Reset()
				m.showSelector = false
				m.selectorCursor = 0

				if strings.HasPrefix(v, "/") {
					m.handleSlashCommand(v)
					return m, nil
				}

				m.addUserMessage(v)

				if content, ok := memory.DetectSaveIntent(v); ok {
					store := memory.Load()
					if err := store.Add(content); err != nil {
						m.addSystemMessage(fmt.Sprintf("Failed to save memory: %v", err))
					} else {
						m.addSystemMessage(fmt.Sprintf("Saved to memory: %q", content))
					}
				}

				m.loading = true
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
			m.err = nil
			if strings.TrimSpace(msg.Text) != "" {
				m.addAssistantMessage(msg.Text)
			}
		}
		m.updateViewport()
	}

	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	inputVal := m.input.Value()
	mode, items := parseInput(inputVal)
	prevVisible := m.showSelector
	if len(items) > 0 {
		m.selectorItems = items
		m.selectorModeState = mode
		m.showSelector = true
		if m.selectorCursor >= len(items) {
			m.selectorCursor = len(items) - 1
		}
	} else {
		m.showSelector = false
		m.selectorItems = nil
	}
	if prevVisible != m.showSelector {
		m.recalcViewportHeight()
	}

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
			m.addSystemMessage("Usage: /agent <provider> (gemini, openai, anthropic, groq, ollama)")
		} else {
			os.Setenv("AI_PROVIDER", arg)
			persistEnv("AI_PROVIDER", arg)
			m.addSystemMessage(fmt.Sprintf("AI provider set to %q", arg))
		}

	case "/mode":
		if arg == "" {
			m.addSystemMessage("Usage: /mode <model> (for example gemini-2.5-flash or gpt-5-mini)")
		} else {
			os.Setenv("AI_MODEL", arg)
			persistEnv("AI_MODEL", arg)
			m.addSystemMessage(fmt.Sprintf("AI model set to %q", arg))
		}

	case "/key":
		if arg == "" {
			m.addSystemMessage("Usage: /key <provider> <api_key>")
		} else {
			keyParts := strings.SplitN(arg, " ", 2)
			if len(keyParts) < 2 {
				m.addSystemMessage("Usage: /key <provider> <api_key>")
			} else {
				provider := strings.ToLower(keyParts[0])
				apiKey := strings.TrimSpace(keyParts[1])
				envKey := providerToEnvKey(provider)
				if envKey == "" {
					m.addSystemMessage(fmt.Sprintf("Unknown provider %q. Use gemini, openai, anthropic, or groq.", provider))
				} else {
					os.Setenv(envKey, apiKey)
					persistEnv(envKey, apiKey)
					m.addSystemMessage(fmt.Sprintf("%s set to %s", envKey, maskKey(apiKey)))
				}
			}
		}

	case "/prompt":
		if arg == "" {
			m.addSystemMessage("Usage: /prompt <name> (default, coder, writer, analyst)")
		} else {
			if _, ok := prompts.GetPrompt(arg); ok {
				os.Setenv("SYSTEM_PROMPT", arg)
				persistEnv("SYSTEM_PROMPT", arg)
				m.addSystemMessage(fmt.Sprintf("System prompt set to %q", arg))
			} else {
				m.addSystemMessage(fmt.Sprintf("Unknown prompt %q. Use /prompts to see options.", arg))
			}
		}

	case "/prompts":
		var sb strings.Builder
		sb.WriteString("Available system prompts:\n")
		active := prompts.GetActivePrompt()
		for _, p := range prompts.ListPrompts() {
			marker := "  "
			if p.Name == active.Name {
				marker = "> "
			}
			sb.WriteString(fmt.Sprintf("%s%s - %s\n", marker, p.Name, p.Description))
		}
		m.addSystemMessage(sb.String())

	case "/ollama-models":
		models, err := agentMod.ListOllamaModels()
		if err != nil {
			m.addSystemMessage(fmt.Sprintf("Failed to list Ollama models: %v", err))
		} else if len(models) == 0 {
			m.addSystemMessage("No local Ollama models found. Pull one with: ollama pull <model>")
		} else {
			var sb strings.Builder
			sb.WriteString("Local Ollama models:\n")
			for _, name := range models {
				sb.WriteString(fmt.Sprintf("  - %s\n", name))
			}
			m.addSystemMessage(sb.String())
		}

	case "/status":
		provider := os.Getenv("AI_PROVIDER")
		if provider == "" {
			provider = "gemini (default)"
		}
		aiModel := os.Getenv("AI_MODEL")
		if aiModel == "" {
			aiModel = "(provider default)"
		}
		activePrompt := prompts.GetActivePrompt()
		ollamaHost := os.Getenv("OLLAMA_HOST")
		if ollamaHost == "" {
			ollamaHost = "http://localhost:11434"
		}
		memStore := memory.Load()

		status := fmt.Sprintf("Current configuration:\n  Session: %s\n  Provider: %s\n  Model: %s\n  System Prompt: %s\n  Chat Entries: %d\n  Memories: %d\n  Computer Control: %s\n  Ollama Host: %s",
			m.sessionID, provider, aiModel, activePrompt.Name, m.chat.HistoryCount(), memStore.Count(), m.permissionMode, ollamaHost)
		m.addSystemMessage(status)

	case "/computer":
		if arg == "" {
			m.addSystemMessage("Usage: /computer <off|workspace|full>")
			break
		}
		mode := computer.ParsePermissionMode(arg)
		switch mode {
		case computer.PermissionOff:
			m.permissionMode = mode
			m.addSystemMessage("Computer control disabled.")
		case computer.PermissionWorkspace:
			m.permissionMode = mode
			m.addSystemMessage("Computer control enabled in workspace mode. GlAgent can run development and inspection commands for you.")
		case computer.PermissionFull:
			m.permissionMode = mode
			m.addSystemMessage("Warning: full computer control is enabled. The agent can now run broader shell commands on this machine. Use it only when you trust the task.")
		default:
			m.addSystemMessage("Usage: /computer <off|workspace|full>")
		}
		_ = m.saveSession()

	case "/session":
		m.addSystemMessage(fmt.Sprintf("Current session: %s\nResume later with: glagent --continue %s", m.sessionID, m.sessionID))

	case "/clear":
		m.chat.Clear()
		m.messages = nil
		m.err = nil
		m.addSystemMessage("Chat history cleared.")

	case "/save":
		if arg == "" {
			m.addSystemMessage("Usage: /save <content to remember>")
		} else {
			store := memory.Load()
			if err := store.Add(arg); err != nil {
				m.addSystemMessage(fmt.Sprintf("Failed to save: %v", err))
			} else {
				m.addSystemMessage(fmt.Sprintf("Saved to memory: %q (%d total)", arg, store.Count()))
			}
		}

	case "/memory":
		store := memory.Load()
		items := store.List()
		if len(items) == 0 {
			m.addSystemMessage("Memory is empty. Use /save or say \"remember that ...\" to add items.")
		} else {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Saved memories (%d items):\n", len(items)))
			for i, item := range items {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, item.Content))
			}
			sb.WriteString("\nUse /forget <number> to remove, or /forget-all to clear.")
			m.addSystemMessage(sb.String())
		}

	case "/forget":
		if arg == "" {
			m.addSystemMessage("Usage: /forget <number> (use /memory to see the numbered list)")
		} else {
			var idx int
			if _, err := fmt.Sscanf(arg, "%d", &idx); err != nil || idx < 1 {
				m.addSystemMessage("Please provide a valid memory number. Use /memory to see the list.")
			} else {
				store := memory.Load()
				if err := store.Remove(idx - 1); err != nil {
					m.addSystemMessage(fmt.Sprintf("Error: %v", err))
				} else {
					m.addSystemMessage(fmt.Sprintf("Memory #%d removed. (%d remaining)", idx, store.Count()))
				}
			}
		}

	case "/forget-all":
		store := memory.Load()
		if err := store.Clear(); err != nil {
			m.addSystemMessage(fmt.Sprintf("Failed to clear memory: %v", err))
		} else {
			m.addSystemMessage("All memories cleared.")
		}

	case "/help":
		var sb strings.Builder
		sb.WriteString("Available commands:\n")
		for _, c := range allCommands {
			sb.WriteString(fmt.Sprintf("  %s - %s\n", c.Name, c.Description))
		}
		sb.WriteString("\nTips:\n  - Type / to see command autocomplete\n  - Use Up/Down to navigate and Tab/Enter to select\n  - Alt+Enter for multi-line input\n  - Use /computer full only when you really want broad shell control")
		m.addSystemMessage(sb.String())

	default:
		m.addSystemMessage(fmt.Sprintf("Unknown command: %s. Type /help for a list of commands.", cmd))
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
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "gemini"
	}
	activePrompt := prompts.GetActivePrompt()
	configInfo := statusBarStyle.Render(fmt.Sprintf("  %s | %s | %s | %s", provider, activePrompt.Name, m.permissionMode, m.sessionID))
	header := fmt.Sprintf("%s%s\n\n", title, configInfo)

	vp := viewportStyle.Render(m.viewport.View())

	status := statusReadyStyle.Render(" Ready")
	if m.loading {
		status = statusLoadingStyle.Render(fmt.Sprintf(" %s Working...", m.spinner.View()))
	}
	if m.err != nil {
		status = errorStyle.Render(fmt.Sprintf(" Error: %v", m.err))
	}

	selectorView := ""
	if m.showSelector && len(m.selectorItems) > 0 {
		selectorView = renderSelector(m.selectorItems, m.selectorCursor, m.width)
	}

	inStyle := inputStyle
	if m.showSelector {
		inStyle = inputActiveStyle
	}
	in := inStyle.Render(m.input.View())

	var parts []string
	parts = append(parts, header, vp, status)
	if selectorView != "" {
		parts = append(parts, selectorView)
	}
	parts = append(parts, in)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m *model) updateViewport() {
	var b strings.Builder
	for _, msg := range m.messages {
		switch msg.Role {
		case "User":
			b.WriteString(userMsgStyle.Render("You: "))
			b.WriteString(msg.Content)
		case "System":
			b.WriteString(sysMsgStyle.Render("* " + msg.Content))
		default:
			b.WriteString(botMsgStyle.Render("GlAgent: "))

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
	chat := m.chat
	permissionMode := m.permissionMode

	return func() tea.Msg {
		text, err := runAgentTurn(chat, prompt, permissionMode)
		return aiResponseMsg{Text: text, Err: err}
	}
}

func runAgentTurn(chat *agentMod.ChatSession, userPrompt string, permissionMode computer.PermissionMode) (string, error) {
	turnPrompt := userPrompt
	finalText := ""

	for step := 0; step < maxAgentSteps; step++ {
		response, err := agentMod.AskAIWithHistoryAndSystem(chat, turnPrompt, computer.Instructions(permissionMode))
		if err != nil {
			return "", err
		}

		commands, cleaned := computer.ExtractCommands(response)
		if len(commands) == 0 {
			return strings.TrimSpace(cleaned), nil
		}

		if !permissionMode.AllowsExecution() {
			if cleaned == "" {
				return "Command execution is disabled. Enable it with /computer workspace or /computer full.", nil
			}
			return cleaned + "\n\nCommand execution is disabled. Enable it with /computer workspace or /computer full.", nil
		}

		if cleaned != "" && finalText == "" {
			finalText = cleaned
		}

		for _, command := range commands {
			result, err := computer.Execute(command.Command, permissionMode, 30*time.Second)
			if err != nil {
				chat.AddSystemMessage(fmt.Sprintf("Command execution failed for %q: %v", command.Command, err))
				return "", err
			}
			chat.AddSystemMessage(computer.FormatResult(result))
		}

		turnPrompt = "The command result has been added to the conversation. Use that real output and answer the user directly. If you still need another command, emit another <glagent_command> block."
	}

	if finalText != "" {
		return finalText, nil
	}
	return "I hit the maximum command-execution steps for one turn. Please narrow the task or ask me to continue.", nil
}

func (m *model) recalcViewportHeight() {
	if m.width == 0 || m.height == 0 {
		return
	}
	chrome := 12
	if m.showSelector && len(m.selectorItems) > 0 {
		chrome += len(m.selectorItems) + 2
	}
	vpHeight := m.height - chrome
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.SetHeight(vpHeight)
}

func (m *model) completeSelection() {
	if len(m.selectorItems) == 0 {
		return
	}

	selected := m.selectorItems[m.selectorCursor]

	if m.selectorModeState == modeCommand {
		m.input.Reset()
		m.input.InsertString(selected.Value + " ")
	} else {
		val := m.input.Value()
		lastSpace := strings.LastIndex(val, " ")
		prefix := ""
		if lastSpace >= 0 {
			prefix = val[:lastSpace+1]
		}
		m.input.Reset()
		m.input.InsertString(prefix + selected.Value)
	}

	m.showSelector = false
	m.selectorCursor = 0
	m.recalcViewportHeight()
}

func (m *model) addUserMessage(content string) {
	msg := message{Role: "User", Content: content, Time: time.Now()}
	m.messages = append(m.messages, msg)
	m.chat.AddUserMessage(content)
	_ = m.saveSession()
}

func (m *model) addAssistantMessage(content string) {
	msg := message{Role: "Assistant", Content: content, Time: time.Now()}
	m.messages = append(m.messages, msg)
	m.chat.AddAssistantMessage(content)
	_ = m.saveSession()
}

func (m *model) addSystemMessage(content string) {
	msg := message{Role: "System", Content: content, Time: time.Now()}
	m.messages = append(m.messages, msg)
	m.chat.AddSystemMessage(content)
	_ = m.saveSession()
}

func (m *model) saveSession() error {
	stored := &sessionstore.Session{
		ID:             m.sessionID,
		Messages:       make([]sessionstore.Message, 0, len(m.messages)),
		ChatEntries:    append([]agentMod.ChatEntry{}, m.chat.Entries...),
		PermissionMode: m.permissionMode.String(),
	}

	for _, msg := range m.messages {
		stored.Messages = append(stored.Messages, sessionstore.Message{
			Role:    msg.Role,
			Content: msg.Content,
			Time:    msg.Time,
		})
	}

	return sessionstore.Save(stored)
}

func persistEnv(key, value string) {
	envMap, err := godotenv.Read(".env")
	if err != nil {
		envMap = make(map[string]string)
	}
	envMap[key] = value
	_ = godotenv.Write(envMap, ".env")
}

func providerToEnvKey(provider string) string {
	switch provider {
	case "gemini", "google":
		return "GOOGLE_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	default:
		return ""
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

package glagentgui

import (
	"os"
	"strings"

	"glagent/src/modules/agentMod"
	"glagent/src/prompts"

	"charm.land/lipgloss/v2"
)

type slashCommand struct {
	Name        string
	Description string
	Usage       string
}

type selectorItem struct {
	Value       string
	Description string
}

type selectorMode int

const (
	modeCommand selectorMode = iota
	modeArgument
)

var allCommands = []slashCommand{
	{Name: "/agent", Description: "Switch AI provider", Usage: "/agent <provider>"},
	{Name: "/mode", Description: "Switch AI model", Usage: "/mode <model>"},
	{Name: "/key", Description: "Set API key for a provider", Usage: "/key <provider> <api_key>"},
	{Name: "/prompt", Description: "Set active system prompt", Usage: "/prompt <name>"},
	{Name: "/prompts", Description: "List available prompts", Usage: "/prompts"},
	{Name: "/computer", Description: "Set computer control mode", Usage: "/computer <off|workspace|full>"},
	{Name: "/session", Description: "Show the current chat session id", Usage: "/session"},
	{Name: "/approvals", Description: "List pending risky actions", Usage: "/approvals"},
	{Name: "/approve", Description: "Approve one pending action", Usage: "/approve <id>"},
	{Name: "/deny", Description: "Deny one pending action", Usage: "/deny <id>"},
	{Name: "/git-status", Description: "Show git status", Usage: "/git-status"},
	{Name: "/git-diff", Description: "Show git diff for a path", Usage: "/git-diff <path>"},
	{Name: "/git-stage", Description: "Stage a path or all current changes", Usage: "/git-stage <path|.>"},
	{Name: "/git-commit", Description: "Create a git commit", Usage: "/git-commit <message>"},
	{Name: "/save", Description: "Save something to memory", Usage: "/save <content>"},
	{Name: "/memory", Description: "View saved memories", Usage: "/memory"},
	{Name: "/forget", Description: "Remove a memory by number", Usage: "/forget <number>"},
	{Name: "/forget-all", Description: "Clear all memories", Usage: "/forget-all"},
	{Name: "/ollama-models", Description: "List local Ollama models", Usage: "/ollama-models"},
	{Name: "/status", Description: "Show current config", Usage: "/status"},
	{Name: "/clear", Description: "Clear chat history", Usage: "/clear"},
	{Name: "/help", Description: "Show help message", Usage: "/help"},
}

var providerSuggestions = []selectorItem{
	{Value: "gemini", Description: "Google Gemini"},
	{Value: "openai", Description: "OpenAI GPT"},
	{Value: "anthropic", Description: "Anthropic Claude"},
	{Value: "groq", Description: "Groq inference"},
	{Value: "ollama", Description: "Local Ollama models"},
}

var keyProviderSuggestions = []selectorItem{
	{Value: "gemini", Description: "GOOGLE_API_KEY"},
	{Value: "openai", Description: "OPENAI_API_KEY"},
	{Value: "anthropic", Description: "ANTHROPIC_API_KEY"},
	{Value: "groq", Description: "GROQ_API_KEY"},
}

var computerModeSuggestions = []selectorItem{
	{Value: "off", Description: "Do not let the agent run commands"},
	{Value: "workspace", Description: "Run dev and inspection commands in this project"},
	{Value: "full", Description: "Broad shell control. Use with care"},
}

var geminiModels = []selectorItem{
	{Value: "gemini-2.5-flash", Description: "Fast and capable"},
	{Value: "gemini-2.5-pro", Description: "Most capable Gemini"},
	{Value: "gemini-2.0-flash", Description: "Previous generation fast"},
	{Value: "gemini-2.0-flash-lite", Description: "Lightweight and cheap"},
	{Value: "gemini-1.5-pro", Description: "Legacy pro"},
}

var openaiModels = []selectorItem{
	{Value: "gpt-5", Description: "Most capable GPT"},
	{Value: "gpt-5-mini", Description: "Fast and affordable"},
	{Value: "gpt-4.1", Description: "Smart and fast"},
	{Value: "gpt-4o", Description: "Multimodal"},
	{Value: "o4-mini", Description: "Fast reasoning"},
	{Value: "o3", Description: "Advanced reasoning"},
}

var anthropicModels = []selectorItem{
	{Value: "claude-opus-4-6", Description: "Most capable Claude"},
	{Value: "claude-sonnet-4-6", Description: "Balanced default"},
	{Value: "claude-3-5-sonnet-latest", Description: "Legacy Sonnet"},
	{Value: "claude-3-5-haiku-latest", Description: "Fast and light"},
}

var groqModels = []selectorItem{
	{Value: "openai/gpt-oss-20b", Description: "Default Groq model"},
	{Value: "llama-3.3-70b-versatile", Description: "Llama 3.3 70B"},
	{Value: "llama-3.1-8b-instant", Description: "Fast Llama"},
	{Value: "mixtral-8x7b-32768", Description: "Mixtral 8x7B"},
	{Value: "deepseek-r1-distill-llama-70b", Description: "DeepSeek R1"},
}

func getModelSuggestions() []selectorItem {
	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "gemini"
	}

	switch provider {
	case "gemini":
		return geminiModels
	case "openai":
		return openaiModels
	case "anthropic":
		return anthropicModels
	case "groq":
		return groqModels
	case "ollama":
		return getOllamaModelSuggestions()
	default:
		return nil
	}
}

func getOllamaModelSuggestions() []selectorItem {
	models, err := agentMod.ListOllamaModels()
	if err == nil && len(models) > 0 {
		var items []selectorItem
		for _, name := range models {
			items = append(items, selectorItem{
				Value:       name,
				Description: "Local model",
			})
		}
		return items
	}

	return []selectorItem{
		{Value: "qwen3:8b", Description: "Qwen 3 8B"},
		{Value: "llama3.1:8b", Description: "Llama 3.1 8B"},
		{Value: "gemma2:9b", Description: "Gemma 2 9B"},
		{Value: "mistral:7b", Description: "Mistral 7B"},
		{Value: "qwen2.5-coder:7b", Description: "Qwen coder"},
	}
}

func getPromptSuggestions() []selectorItem {
	var items []selectorItem
	for _, p := range prompts.ListPrompts() {
		items = append(items, selectorItem{
			Value:       p.Name,
			Description: p.Description,
		})
	}
	return items
}

func parseInput(input string) (selectorMode, []selectorItem) {
	if !strings.HasPrefix(input, "/") || strings.Contains(input, "\n") {
		return modeCommand, nil
	}

	parts := strings.SplitN(input, " ", 3)
	if len(parts) == 1 {
		return modeCommand, filterCommandsAsItems(input)
	}

	cmdName := parts[0]
	currentArg := ""
	argIndex := 0

	if len(parts) == 2 {
		currentArg = parts[1]
	} else {
		currentArg = parts[2]
		argIndex = 1
	}

	suggestions := getArgumentSuggestions(cmdName, argIndex, currentArg)
	if len(suggestions) > 0 {
		return modeArgument, suggestions
	}

	return modeArgument, nil
}

func getArgumentSuggestions(command string, argIndex int, currentArg string) []selectorItem {
	var pool []selectorItem

	switch command {
	case "/agent":
		if argIndex == 0 {
			pool = providerSuggestions
		}
	case "/mode":
		if argIndex == 0 {
			pool = getModelSuggestions()
		}
	case "/key":
		if argIndex == 0 {
			pool = keyProviderSuggestions
		}
	case "/prompt":
		if argIndex == 0 {
			pool = getPromptSuggestions()
		}
	case "/computer":
		if argIndex == 0 {
			pool = computerModeSuggestions
		}
	}

	if pool == nil {
		return nil
	}

	return filterItems(pool, currentArg)
}

func filterCommandsAsItems(prefix string) []selectorItem {
	prefix = strings.ToLower(prefix)
	var items []selectorItem
	for _, cmd := range allCommands {
		if strings.HasPrefix(strings.ToLower(cmd.Name), prefix) {
			items = append(items, selectorItem{
				Value:       cmd.Name,
				Description: cmd.Description,
			})
		}
	}
	return items
}

func filterItems(items []selectorItem, prefix string) []selectorItem {
	if prefix == "" {
		return items
	}
	prefix = strings.ToLower(prefix)
	var filtered []selectorItem
	for _, item := range items {
		if strings.HasPrefix(strings.ToLower(item.Value), prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

var (
	selectorBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#B07CD8")).
				Padding(0, 1)

	selectorItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#D4BEE4")).
				PaddingLeft(1)

	selectorActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1A1A2E")).
				Background(lipgloss.Color("#FF8C42")).
				Bold(true).
				PaddingLeft(1).
				PaddingRight(1)

	selectorDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8B7DA8")).
				Italic(true)

	selectorActiveDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1A1A2E")).
				Background(lipgloss.Color("#FF8C42")).
				Italic(true)
)

func renderSelector(items []selectorItem, cursor int, width int) string {
	if len(items) == 0 {
		return ""
	}

	var b strings.Builder
	for i, item := range items {
		nameStr := item.Value
		descStr := " - " + item.Description

		if i == cursor {
			line := selectorActiveStyle.Render(nameStr) + selectorActiveDescStyle.Render(descStr)
			b.WriteString(line)
		} else {
			name := selectorItemStyle.Render(nameStr)
			desc := selectorDescStyle.Render(descStr)
			b.WriteString(name + desc)
		}

		if i < len(items)-1 {
			b.WriteString("\n")
		}
	}

	return selectorBorderStyle.Width(width - 4).Render(b.String())
}

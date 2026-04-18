package agentMod

import (
	"errors"
	"os"
	"strings"

	"glagent/src/modules/agentMod/providers"
	consolemarkdown "glagent/src/modules/consoleMarkdown"
	"glagent/src/modules/memory"
	"glagent/src/prompts"

	"github.com/joho/godotenv"
)

// buildSystemPrompt combines the active system prompt with any saved memories.
func buildSystemPrompt(extraParts ...string) string {
	base := prompts.GetActivePrompt().Content
	store := memory.Load()
	memCtx := store.BuildContext()

	parts := []string{base}
	if memCtx != "" {
		parts = append(parts, memCtx)
	}
	for _, part := range extraParts {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n\n")
}

func getProvider() (providers.Provider, string, error) {
	_ = godotenv.Load()

	name := os.Getenv("AI_PROVIDER")
	model := os.Getenv("AI_MODEL")

	switch name {
	case "gemini", "":
		return providers.NewGeminiProvider(), model, nil
	case "openai":
		return providers.NewOpenAIProvider(), model, nil
	case "anthropic":
		p, err := providers.NewAnthropicProvider()
		if err != nil {
			return nil, "", err
		}
		return p, model, nil
	case "groq":
		return providers.NewGroqProvider(), model, nil
	case "ollama":
		return providers.NewOllamaProvider(), model, nil
	default:
		return nil, "", errors.New("unsupported AI_PROVIDER: " + name)
	}
}

// AskAI sends a prompt to the active AI provider with the active system prompt.
func AskAI(prompt string) (string, error) {
	return AskAIWithSystem(prompt, "")
}

func AskAIWithSystem(prompt string, extraSystem string) (string, error) {
	provider, model, err := getProvider()
	if err != nil {
		return "", err
	}

	systemPrompt := buildSystemPrompt(extraSystem)

	return provider.Generate(prompt, providers.GenerateOptions{
		Model:        model,
		SystemPrompt: systemPrompt,
	})
}

// AskAIWithHistory sends a prompt with conversation context to the active AI provider.
func AskAIWithHistory(session *ChatSession, userInput string) (string, error) {
	return AskAIWithHistoryAndSystem(session, userInput, "")
}

func AskAIWithHistoryAndSystem(session *ChatSession, userInput string, extraSystem string) (string, error) {
	provider, model, err := getProvider()
	if err != nil {
		return "", err
	}

	systemPrompt := buildSystemPrompt(extraSystem)
	contextPrompt := session.BuildPrompt(userInput)

	return provider.Generate(contextPrompt, providers.GenerateOptions{
		Model:        model,
		SystemPrompt: systemPrompt,
	})
}

func AskAIStream(prompt string) (string, error) {
	return AskAIStreamWithSystem(prompt, "")
}

func AskAIStreamWithSystem(prompt string, extraSystem string) (string, error) {
	provider, model, err := getProvider()
	if err != nil {
		return "", err
	}

	systemPrompt := buildSystemPrompt(extraSystem)

	var full string

	out, err := provider.GenerateStream(prompt, providers.GenerateOptions{
		Model:        model,
		SystemPrompt: systemPrompt,
	}, func(token string) {
		full += token
		consolemarkdown.LiveRenderMarkdown(full)
	})

	if err != nil {
		return "", err
	}

	consolemarkdown.LiveRenderMarkdown(out)
	consolemarkdown.FinishLiveRender()

	return out, nil
}

// ListOllamaModels returns locally available Ollama models.
func ListOllamaModels() ([]string, error) {
	p := providers.NewOllamaProvider()
	return p.ListModels()
}

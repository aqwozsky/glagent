package agentMod

import (
	"errors"
	"os"

	"glagent/src/modules/agentMod/providers"
	consolemarkdown "glagent/src/modules/consoleMarkdown"

	"github.com/joho/godotenv"
)

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

func AskAI(prompt string) (string, error) {
	provider, model, err := getProvider()
	if err != nil {
		return "", err
	}

	return provider.Generate(prompt, providers.GenerateOptions{
		Model: model,
	})
}

func AskAIStream(prompt string) (string, error) {
	provider, model, err := getProvider()
	if err != nil {
		return "", err
	}

	var full string

	out, err := provider.GenerateStream(prompt, providers.GenerateOptions{
		Model: model,
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
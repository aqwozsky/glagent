package providers

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

type GroqProvider struct{}

func NewGroqProvider() *GroqProvider {
	return &GroqProvider{}
}

func (p *GroqProvider) Name() string {
	return "groq"
}

func (p *GroqProvider) Generate(prompt string, opts GenerateOptions) (string, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", errors.New("GROQ_API_KEY not found")
	}

	model := opts.Model
	if model == "" {
		model = "openai/gpt-oss-20b"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.groq.com/openai/v1"),
	)

	ctx := context.Background()

	input := prompt
	if opts.SystemPrompt != "" {
		input = opts.SystemPrompt + "\n\n" + prompt
	}

	resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(input),
		},
	})
	if err != nil {
		return "", err
	}

	out := strings.TrimSpace(resp.OutputText())
	if out == "" {
		return "", errors.New("empty response")
	}

	return out, nil
}

func (p *GroqProvider) GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error) {
	out, err := p.Generate(prompt, opts)
	if err != nil {
		return "", err
	}
	if onToken != nil {
		onToken(out)
	}
	return out, nil
}
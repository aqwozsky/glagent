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

type OpenAIProvider struct{}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Generate(prompt string, opts GenerateOptions) (string, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY not found")
	}

	model := opts.Model
	if model == "" {
		model = "gpt-5-mini"
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
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

func (p *OpenAIProvider) GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error) {
	// İlk sürümde stream yerine normal çağrı yapıp token callback simüle edebilirsin.
	// Sonra SSE stream eklenir.
	out, err := p.Generate(prompt, opts)
	if err != nil {
		return "", err
	}
	if onToken != nil {
		onToken(out)
	}
	return out, nil
}
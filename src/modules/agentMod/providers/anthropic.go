package providers

import (
	"context"
	"errors"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
)

type AnthropicProvider struct {
	client anthropic.Client
}

func NewAnthropicProvider() (*AnthropicProvider, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY not set")
	}

	return &AnthropicProvider{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
		),
	}, nil
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

func (p *AnthropicProvider) Generate(prompt string, opts GenerateOptions) (string, error) {
	model := opts.Model
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	params := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	if opts.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: opts.SystemPrompt},
		}
	}

	resp, err := p.client.Messages.New(context.Background(), params)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, c := range resp.Content {
		if c.Text != "" {
			sb.WriteString(c.Text)
		}
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", errors.New("empty response")
	}

	return out, nil
}

func (p *AnthropicProvider) GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error) {
	model := opts.Model
	if model == "" {
		model = "claude-sonnet-4-6"
	}

	params := anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: 2048,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	if opts.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: opts.SystemPrompt},
		}
	}

	stream := p.client.Messages.NewStreaming(context.Background(), params)

	var sb strings.Builder
	for stream.Next() {
		event := stream.Current()
		switch ev := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := ev.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				if onToken != nil {
					onToken(delta.Text)
				}
				sb.WriteString(delta.Text)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return "", err
	}

	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", errors.New("empty response")
	}

	return out, nil
}
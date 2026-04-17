package providers

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

type GeminiProvider struct{}

func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{}
}

func (p *GeminiProvider) Name() string {
	return "gemini"
}

func (p *GeminiProvider) Generate(prompt string, opts GenerateOptions) (string, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", errors.New("GOOGLE_API_KEY not found")
	}

	model := opts.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	var contents string
	if opts.SystemPrompt != "" {
		contents = opts.SystemPrompt + "\n\n" + prompt
	} else {
		contents = prompt
	}

	resp, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(contents),
		nil,
	)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", errors.New("nil response")
	}

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		return "", errors.New("empty response")
	}

	return text, nil
}

func (p *GeminiProvider) GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error) {
	_ = godotenv.Load()

	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return "", errors.New("GOOGLE_API_KEY not found")
	}

	model := opts.Model
	if model == "" {
		model = "gemini-2.5-flash"
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", err
	}

	var contents string
	if opts.SystemPrompt != "" {
		contents = opts.SystemPrompt + "\n\n" + prompt
	} else {
		contents = prompt
	}

	var full strings.Builder

	for result, err := range client.Models.GenerateContentStream(
		ctx,
		model,
		genai.Text(contents),
		nil,
	) {
		if err != nil {
			return "", err
		}

		if result == nil || len(result.Candidates) == 0 || result.Candidates[0] == nil || result.Candidates[0].Content == nil {
			continue
		}

		for _, part := range result.Candidates[0].Content.Parts {
			if part == nil || part.Text == "" {
				continue
			}

			full.WriteString(part.Text)
			if onToken != nil {
				onToken(part.Text)
			}
		}
	}

	out := strings.TrimSpace(full.String())
	if out == "" {
		return "", errors.New("empty response")
	}

	return out, nil
}
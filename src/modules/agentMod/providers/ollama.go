package providers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type OllamaProvider struct{}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func (p *OllamaProvider) Generate(prompt string, opts GenerateOptions) (string, error) {
	model := opts.Model
	if model == "" {
		model = "qwen3:8b"
	}

	finalPrompt := prompt
	if opts.SystemPrompt != "" {
		finalPrompt = opts.SystemPrompt + "\n\n" + prompt
	}

	body, _ := json.Marshal(ollamaRequest{
		Model:  model,
		Prompt: finalPrompt,
		Stream: false,
	})

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}

	text := strings.TrimSpace(out.Response)
	if text == "" {
		return "", errors.New("empty response")
	}

	return text, nil
}

func (p *OllamaProvider) GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error) {
	out, err := p.Generate(prompt, opts)
	if err != nil {
		return "", err
	}
	if onToken != nil {
		onToken(out)
	}
	return out, nil
}
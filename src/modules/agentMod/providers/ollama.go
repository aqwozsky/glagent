package providers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

type OllamaProvider struct{}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{}
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) baseURL() string {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	return strings.TrimRight(host, "/")
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ollamaTagsResponse struct {
	Models []ollamaModelInfo `json:"models"`
}

type ollamaModelInfo struct {
	Name string `json:"name"`
}

// ListModels calls /api/tags and returns the list of locally available models.
func (p *OllamaProvider) ListModels() ([]string, error) {
	resp, err := http.Get(p.baseURL() + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	var names []string
	for _, m := range tags.Models {
		names = append(names, m.Name)
	}
	return names, nil
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

	resp, err := http.Post(p.baseURL()+"/api/generate", "application/json", bytes.NewBuffer(body))
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
		Stream: true,
	})

	resp, err := http.Post(p.baseURL()+"/api/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var full strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var chunk ollamaResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}

		if chunk.Response != "" {
			full.WriteString(chunk.Response)
			if onToken != nil {
				onToken(chunk.Response)
			}
		}

		if chunk.Done {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	out := strings.TrimSpace(full.String())
	if out == "" {
		return "", errors.New("empty response")
	}

	return out, nil
}
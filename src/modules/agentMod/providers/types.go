package providers

type GenerateOptions struct {
	SystemPrompt string
	Model        string
}

type Provider interface {
	Name() string
	Generate(prompt string, opts GenerateOptions) (string, error)
	GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error)
}
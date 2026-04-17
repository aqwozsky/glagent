# GlAgent

GlAgent is a powerful, extensible command-line AI assistant built in Go. It supports multiple LLM providers (Gemini, OpenAI, Anthropic, Groq, Ollama) and features a modern, interactive Terminal UI (TUI) powered by Charm's [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features
- **Multi-Provider Support:** Easily switch between different AI models and providers.
- **Modern CLUI:** A sleek, interactive console GUI utilizing `bubbletea` and `lipgloss` for a premium user experience.
- **Markdown Rendering:** Beautiful inline markdown rendering for AI responses in your terminal.
- **Session History:** Keeps track of conversational context with customizable max turns.

## Installation

1. Clone the repository:
   ```bash
   git clone <your-repo-url>
   cd GlAgent
   ```
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up environment variables. Create a `.env` file in the root directory:
   ```env
   AI_PROVIDER=gemini
   AI_MODEL=gemini-2.5-flash
   GEMINI_API_KEY=your-api-key
   
   # Other providers require their respective keys:
   # OPENAI_API_KEY=
   # ANTHROPIC_API_KEY=
   # GROQ_API_KEY=
   ```

## Usage

Run the agent from the project root:
```bash
go run main.go
```

Use `Tab` to navigate, and type your prompt. Press `Enter` to send your message to the agent.

# GlAgent

GlAgent is a terminal-first AI agent built in Go. It combines a Bubble Tea TUI, multiple LLM providers, persistent chat sessions, lightweight memory, and local command execution so it can both talk about work and actually do parts of it.

![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat-square&logo=go)
![UI](https://img.shields.io/badge/UI-Bubble%20Tea-ffb347?style=flat-square)
![Providers](https://img.shields.io/badge/LLMs-Gemini%20%7C%20OpenAI%20%7C%20Anthropic%20%7C%20Groq%20%7C%20Ollama-6c3baa?style=flat-square)

## What It Does

- Chat with multiple AI providers from one terminal UI
- Run local commands when the model needs real output from your machine
- Save and resume chat sessions with `--continue <chat-id>`
- Store lightweight long-term memory in `memory.json`
- Switch system prompts for coding, writing, or analysis
- Use autocomplete for slash commands and common arguments
- Work with cloud models or local Ollama models

## Why It Exists

Many terminal assistants stop at "run this command." GlAgent is designed to go one step further. When you allow it, the model can ask the app to execute a command, read the real result, and then continue the answer using that verified output.

That means prompts like:

- "check my npm version"
- "list local Ollama models"
- "inspect this repo"
- "verify the build"

can become actual actions instead of instructions for you to copy and paste.

## Core Features

### Multi-Provider Support

GlAgent supports:

- Google Gemini
- OpenAI
- Anthropic
- Groq
- Ollama

Provider selection is controlled with:

- environment variables in `.env`
- `/agent <provider>`
- `/mode <model>`

### Agentic Command Execution

GlAgent can run local PowerShell commands and feed their output back into the model. This is how the assistant can verify local facts instead of guessing.

Execution modes:

- `off`: the model cannot run commands
- `workspace`: the model can run normal development and inspection commands
- `full`: the model can run broader machine commands and will warn the user when enabled

Workspace mode is the default because it is the best balance of usefulness and safety for local development work.

### Persistent Sessions

Each chat gets an id like `chat-20260418-153000`. Sessions are written to `.glagent/sessions/<chat-id>.json` and include:

- visible UI messages
- structured chat history used for future prompts
- current computer-control mode
- timestamps

You can resume a chat later with:

```bash
go run main.go --continue chat-20260418-153000
```

### Prompt Presets

Built-in prompt modes:

- `default`
- `coder`
- `writer`
- `analyst`

These shape the assistant's behavior and are combined with memory and execution rules before each model call.

### Memory

GlAgent keeps a small persistent memory store in `memory.json`.

You can save memory with:

- `/save <text>`
- natural language like `remember that my package manager is pnpm`

Saved memories are injected into the system prompt so the model can keep important user-specific context across sessions.

### TUI Experience

The UI is built with Charm's Bubble Tea ecosystem and includes:

- scrollable conversation viewport
- markdown rendering for model responses
- slash command autocomplete
- provider and prompt switching
- session and permission visibility in the header

## Installation

### Requirements

- Go 1.26 or newer
- PowerShell available on the host machine
- at least one AI provider key, unless you only want local Ollama usage

### Clone

```bash
git clone <your-repo-url>
cd GlAgent
```

### Install Dependencies

```bash
go mod download
```

### Configure Environment

Create a `.env` file in the project root:

```env
AI_PROVIDER=gemini
AI_MODEL=gemini-2.5-flash
SYSTEM_PROMPT=default

GOOGLE_API_KEY=your-google-key
# OPENAI_API_KEY=
# ANTHROPIC_API_KEY=
# GROQ_API_KEY=

# Optional
# OLLAMA_HOST=http://localhost:11434
```

Provider key mapping:

- `gemini` -> `GOOGLE_API_KEY`
- `openai` -> `OPENAI_API_KEY`
- `anthropic` -> `ANTHROPIC_API_KEY`
- `groq` -> `GROQ_API_KEY`

## Usage

### Start Normally

```bash
go run main.go
```

### Resume a Session

```bash
go run main.go --continue chat-20260418-153000
```

### Start with a Custom Session Id

```bash
go run main.go --session repo-audit
```

## Slash Commands

### Configuration

- `/agent <provider>`: switch provider
- `/mode <model>`: switch model
- `/key <provider> <api_key>`: save a provider API key into `.env`
- `/prompt <name>`: switch system prompt
- `/prompts`: list available prompts
- `/status`: show current provider, model, prompt, session, memory count, and execution mode

### Computer Control

- `/computer off`: disable command execution
- `/computer workspace`: allow project-scoped command execution
- `/computer full`: allow broader shell control and show a warning

### Session

- `/session`: show the current chat id and resume command
- `--continue <chat-id>`: reopen a saved session from the CLI

### Memory

- `/save <text>`: save a memory
- `/memory`: list all saved memories
- `/forget <number>`: delete one memory
- `/forget-all`: clear all memories

### Utilities

- `/ollama-models`: list local Ollama models
- `/clear`: clear current visible chat state
- `/help`: show available commands

## How Command Execution Works

GlAgent does not directly let the model run shell commands. Instead, the app uses a simple protocol:

1. The model receives system instructions telling it how to request execution.
2. If it needs a command, it emits:

```text
<glagent_command>
npm -v
</glagent_command>
```

3. GlAgent extracts the command block.
4. The app runs the command in PowerShell.
5. The real stdout, stderr, exit code, duration, and working directory are added to chat history as a system message.
6. The model gets another turn and answers using the real result.

This makes the assistant much more useful for local verification tasks while keeping execution inside app-controlled code.

## Safety Model

GlAgent uses three permission levels:

- `off`: commands are blocked
- `workspace`: commands are allowed, but obvious destructive patterns are blocked
- `full`: broader execution is allowed

In workspace mode, GlAgent currently blocks obvious destructive command fragments such as:

- `rm`
- `Remove-Item`
- `rmdir`
- `format`
- `shutdown`
- `taskkill`

This is only a lightweight safeguard, not a full sandbox. If you enable `full`, the assistant should be treated as having broad shell-level power on the machine.

## Storage

Files GlAgent writes today:

- `.env`: provider, model, and prompt configuration
- `memory.json`: saved memories
- `.glagent/sessions/*.json`: persisted chat sessions

## Project Structure

```text
main.go
src/
  modules/
    agentMod/
      providers/
    computer/
    consoleMarkdown/
    customLogging/
    glagentGui/
    memory/
    sessionstore/
  prompts/
docs/
  DEVELOPERS.md
```

## Development

Run formatting:

```bash
gofmt -w .
```

Run tests:

```bash
go test ./...
```

Build everything:

```bash
go build ./...
```

For architecture, extension points, storage details, and internal flow, see [docs/DEVELOPERS.md](C:/Users/amesa/Desktop/GlAgent/docs/DEVELOPERS.md).

## Current Limitations

- Command execution currently uses PowerShell directly and is Windows-oriented
- The safety checks are heuristic, not a real sandbox
- Session persistence is file-based and intentionally simple
- There are no automated behavioral tests yet, only compile-level coverage unless you add more tests
- "Full computer control" currently means shell execution, not full desktop GUI automation

## Roadmap Ideas

- richer approval flows for sensitive commands
- first-class desktop automation
- streaming command output in the UI
- better provider/model metadata
- structured tool calls instead of a tag-based protocol
- automated tests around session restore and command execution loops

## License

See [LICENCE](C:/Users/amesa/Desktop/GlAgent/LICENCE).

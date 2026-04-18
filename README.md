# GlAgent

GlAgent is a terminal-first AI agent built in Go. It combines a Bubble Tea TUI, multiple LLM providers, persistent chat sessions, lightweight memory, and local command execution so it can both talk about work and actually do parts of it.

![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat-square&logo=go)
![UI](https://img.shields.io/badge/UI-Bubble%20Tea-ffb347?style=flat-square)
![Providers](https://img.shields.io/badge/LLMs-Gemini%20%7C%20OpenAI%20%7C%20Anthropic%20%7C%20Groq%20%7C%20Ollama-6c3baa?style=flat-square)

## What It Does

- Chat with multiple AI providers from one terminal UI
- Run local commands when the model needs real output from your machine
- Read, list, patch, move, delete, and write files through built-in structured file actions
- Save and resume chat sessions with `--continue <chat-id>`
- Store lightweight long-term memory in `memory.json`
- Switch between planning and running workflows
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

### Agentic Command and File Execution

GlAgent can run local PowerShell commands and also perform first-class file operations. This is how the assistant can verify local facts, inspect project files, and save changes without only falling back to "run this yourself."

Execution modes:

- `off`: the model cannot run commands
- `workspace`: the model can run normal development and inspection commands
- `full`: the model can run broader machine commands and will warn the user when enabled

Workspace mode is the default because it is the best balance of usefulness and safety for local development work.

Built-in file actions:

- read a file
- list a directory
- write or replace a file
- append to a file
- patch exact text inside a file
- create a directory
- move or rename a file
- delete a file or directory

In `workspace` mode, file actions are limited to the current project directory. In `full` mode, broader file access is allowed.
Risky actions can pause for approval before execution.

### Planning and Running Modes

GlAgent now separates workflow mode from model selection:

- `run`: the default mode, where GlAgent can inspect, execute, edit, and verify
- `plan`: a planning mode, where GlAgent should stop before running commands or changing files and instead return a concrete next-step plan

Switch with:

- `/workflow run`
- `/workflow plan`

Workflow mode is saved with the session, so `--continue` restores whether that chat was in planning or running mode.

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

### Optional Windows Setup Command

GlAgent includes a built-in installer flow:

```bash
go run main.go setup
```

The installer prefers copying the current executable, so once GlAgent is installed it can keep working even if you later remove the source checkout.

By default this installs to the current user's standard binary directory:

- Windows: `%LocalAppData%\Programs\GlAgent`
- Linux: `~/.local/bin`

For a machine-wide install into `Program Files`:

```bash
go run main.go setup --system
```

On Linux, `--system` installs into `/usr/local/bin`.

Optional flags:

- `--install-dir <path>`: install to a custom directory
- `--binary-name <name>`: change the installed executable name

Install scripts are also included:

```powershell
./scripts/install.ps1
./scripts/install.ps1 -System
```

```bash
./scripts/install.sh
./scripts/install.sh --system
```

After setup, open a new terminal and run:

```bash
glagent
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

If installed through setup:

```bash
glagent
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
- `/workflow run`: let GlAgent act directly
- `/workflow plan`: keep GlAgent in planning mode

### Approvals

- `/approvals`: list pending risky actions
- `/approve <id>`: approve one risky action
- `/deny <id>`: deny one risky action

Approvals now persist in saved sessions, so if you exit and continue the chat later, pending actions are still there.

### Git

- `/git-status`: show repo status
- `/git-diff <path>`: show diff for a file or path
- `/git-stage <path|.>`: stage one path or all current changes
- `/git-commit <message>`: create a commit from inside GlAgent

### Session

- `/session`: show the current chat id and resume command
- `--continue <chat-id>`: reopen a saved session from the CLI

### Memory

- `/save <text>`: save a memory
- `/memory`: list all saved memories
- `/forget <memory-id>`: delete one memory by id
- `/clear memory`: clear all saved memories
- `/clear all`: clear chat history and all saved memories

### Utilities

- `/ollama-models`: list local Ollama models
- `/clear`: clear current visible chat state
- `/help`: show available commands

## How Execution Works

GlAgent does not directly let the model run shell commands or mutate files freely. Instead, the app exposes simple structured protocols that it parses and executes itself.

### Command Protocol

If the model needs a shell command, it emits:

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

### File Protocol

If the model needs a file read, directory listing, or file write, it can emit:

```text
<glagent_file_read>
README.md
</glagent_file_read>
```

```text
<glagent_file_list>
src/modules
</glagent_file_list>
```

```text
<glagent_file_write path="notes.txt">
hello from glagent
</glagent_file_write>
```

```text
<glagent_file_patch path="README.md">
<<OLD>>
old text
<</OLD>>
<<NEW>>
new text
<</NEW>>
</glagent_file_patch>
```

```text
<glagent_file_move from="old.txt" to="new.txt"></glagent_file_move>
```

```text
<glagent_file_delete path="tmp/output.txt"></glagent_file_delete>
```

The app resolves the target path, enforces workspace restrictions when applicable, performs the operation, and sends the real result back into the conversation as a system message.

This makes the assistant much more useful for local verification and editing tasks while keeping execution inside app-controlled code.

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

For built-in file actions:

- `workspace` mode restricts file paths to the current repo root
- `full` mode allows broader file access
- writes still go through the app's own file-operation layer rather than raw model output being applied directly
- risky rewrites, deletes, renames, patches to sensitive files, installs, and state-changing git commands can be paused for approval
- risky approvals include previews so you can see what is about to change before confirming

## Storage

Files GlAgent writes today:

- `.env`: provider, model, and prompt configuration
- `memory.json`: saved memories
- `.glagent/sessions/*.json`: persisted chat sessions

Saved sessions also include the current workflow mode, permission mode, pending approvals, and structured chat context.

## Project Structure

```text
main.go
src/
  modules/
    agentMod/
      providers/
    computer/
    filesys/
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
- Patch editing currently relies on exact text matches
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

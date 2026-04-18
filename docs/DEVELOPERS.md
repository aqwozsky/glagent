# GlAgent Developer Guide

This document is the implementation-focused companion to the main README. It explains how GlAgent is structured, how data flows through the app, what files are persisted, and where to extend the system safely.

## Goals

GlAgent is built around a few simple goals:

- keep the codebase easy to understand
- make terminal AI useful for real local tasks
- support multiple model providers behind one interface
- preserve context across sessions without a heavy database
- let the app, not the model, own command execution

## High-Level Architecture

```text
main.go
  -> glagentGui.StartGUI(...)
    -> Bubble Tea model
      -> agentMod.AskAIWithHistoryAndSystem(...)
        -> provider adapter
      -> computer.Execute(...)
      -> sessionstore.Save(...)
      -> memory.Load()/Add()/BuildContext()
```

Major layers:

- `main.go`
  CLI entry point and session flag parsing
- `src/modules/glagentGui`
  Bubble Tea UI, slash commands, message rendering, agent loop orchestration
- `src/modules/agentMod`
  prompt composition, provider selection, chat-session prompt building
- `src/modules/agentMod/providers`
  provider-specific API adapters
- `src/modules/computer`
  command protocol, execution permissions, PowerShell process execution
- `src/modules/memory`
  persistent long-term memory store
- `src/modules/sessionstore`
  persisted session files
- `src/prompts`
  built-in prompt presets

## Startup Flow

`main.go` parses:

- `--continue <chat-id>`
- `--session <custom-id>`

Then it calls `glagentGui.StartGUI(options)`.

`StartGUI`:

1. builds the initial Bubble Tea model
2. either restores a stored session or creates a fresh one
3. starts the TUI program

Fresh sessions default to:

- a generated session id from `sessionstore.NewID()`
- `workspace` computer-control mode
- a fresh `agentMod.ChatSession`

## Bubble Tea Model

The main UI model lives in [src/modules/glagentGui/glagentGui.go](C:/Users/amesa/Desktop/GlAgent/src/modules/glagentGui/glagentGui.go).

It owns:

- viewport state
- textarea state
- spinner state
- visible messages
- slash-command selector state
- active `ChatSession`
- session id
- permission mode
- errors

Important methods:

- `Update`
- `View`
- `handleSlashCommand`
- `makeAgentCall`
- `updateViewport`
- `saveSession`

## Message Model vs Chat Model

There are two related but different histories:

### Visible UI Messages

Stored in the GUI model as:

- `User`
- `Assistant`
- `System`

These are what the person sees in the TUI.

### Prompt-Building Chat Entries

Stored in `agentMod.ChatSession` as structured `ChatEntry` values:

- role
- content

These entries are used to build the text prompt sent to the provider.

The split exists so the UI and the AI-context pipeline can evolve independently if needed.

## Agent Loop

The main agent loop is `runAgentTurn(...)` in [src/modules/glagentGui/glagentGui.go](C:/Users/amesa/Desktop/GlAgent/src/modules/glagentGui/glagentGui.go).

Flow:

1. User sends a message.
2. The message is added to visible history and chat history.
3. The app calls `agentMod.AskAIWithHistoryAndSystem(...)`.
4. Extra system instructions from `computer.Instructions(...)` tell the model how to request command execution.
5. If the model returns no command blocks, the response is shown directly.
6. If the model returns one or more `<glagent_command>` blocks:
   the app extracts them with `computer.ExtractCommands(...)`
7. Each command is executed with `computer.Execute(...)`.
8. Results are written back into chat history as system messages.
9. The model gets another turn and answers using the real command result.

The loop is intentionally bounded by `maxAgentSteps` to avoid runaway tool loops.

## Prompt Composition

Prompt composition happens in [src/modules/agentMod/agent.go](C:/Users/amesa/Desktop/GlAgent/src/modules/agentMod/agent.go).

The effective system prompt combines:

1. the active built-in prompt from `src/prompts/prompts.go`
2. saved memory context from `memory.BuildContext()`
3. optional extra system instructions such as command-execution rules

This is done in `buildSystemPrompt(...)`.

## ChatSession Details

`ChatSession` lives in [src/modules/agentMod/chatSession.go](C:/Users/amesa/Desktop/GlAgent/src/modules/agentMod/chatSession.go).

Current behavior:

- stores `[]ChatEntry`
- normalizes roles into `User`, `Assistant`, or `System`
- builds a plain text transcript prompt
- trims history using `MaxTurns`

Current trim logic keeps up to `MaxTurns * 4` entries. That gives room for user, assistant, and additional system messages such as command results.

## Provider Layer

All providers implement the interface in [src/modules/agentMod/providers/types.go](C:/Users/amesa/Desktop/GlAgent/src/modules/agentMod/providers/types.go):

```go
type Provider interface {
    Name() string
    Generate(prompt string, opts GenerateOptions) (string, error)
    GenerateStream(prompt string, opts GenerateOptions, onToken func(string)) (string, error)
}
```

Current adapters:

- `gemini.go`
- `openai.go`
- `anthropic.go`
- `groq.go`
- `ollama.go`

Provider selection happens in `getProvider()` based on:

- `AI_PROVIDER`
- `AI_MODEL`

### Adding a New Provider

To add a provider:

1. create a new file in `src/modules/agentMod/providers`
2. implement `Provider`
3. register it in `getProvider()` inside `agent.go`
4. add it to selector suggestions in `commandSelector.go`
5. add environment-variable mapping in `providerToEnvKey(...)` if needed
6. document it in `README.md`

## Computer Module

The command-execution layer is in [src/modules/computer/computer.go](C:/Users/amesa/Desktop/GlAgent/src/modules/computer/computer.go).

Responsibilities:

- define permission modes
- generate LLM instructions for execution requests
- extract tagged command requests from model output
- execute commands in PowerShell
- format structured execution results
- block obvious dangerous commands in workspace mode

Important types:

- `PermissionMode`
- `CommandRequest`
- `ExecutionResult`

Important functions:

- `Instructions(...)`
- `ExtractCommands(...)`
- `Execute(...)`
- `FormatResult(...)`

### Safety Notes

Current safety is intentionally lightweight. In workspace mode, the app blocks obvious destructive fragments with string matching. This is useful, but it is not equivalent to:

- a sandbox
- policy engine
- file-level allowlist
- per-command approval

If you plan to increase automation power, this is the first subsystem to harden.

## Session Store

Session persistence is implemented in [src/modules/sessionstore/sessionstore.go](C:/Users/amesa/Desktop/GlAgent/src/modules/sessionstore/sessionstore.go).

Sessions are stored under:

```text
.glagent/sessions/<chat-id>.json
```

Current stored fields:

- session id
- created/updated timestamps
- visible messages
- structured chat entries
- permission mode

This design keeps persistence transparent and easy to inspect manually.

### Why File-Based Storage

The current JSON-file approach was chosen because it is:

- simple
- debuggable
- portable
- enough for a single-user local tool

If GlAgent grows into a multi-workspace or multi-user system, this is a natural area to replace with a more structured store.

## Memory Module

Memory lives in [src/modules/memory/memory.go](C:/Users/amesa/Desktop/GlAgent/src/modules/memory/memory.go).

It provides:

- `Load()`
- `Add(...)`
- `Remove(...)`
- `Clear()`
- `List()`
- `Count()`
- `BuildContext()`
- `DetectSaveIntent(...)`

The store is backed by `memory.json`.

Current design is intentionally small and explicit. Memory is treated more like pinned user facts than a long conversation archive.

## Prompt Presets

Prompt presets live in [src/prompts/prompts.go](C:/Users/amesa/Desktop/GlAgent/src/prompts/prompts.go).

Current presets:

- `default`
- `coder`
- `writer`
- `analyst`

### Adding a New Prompt

1. add a new `Prompt` entry to `builtinPrompts`
2. ensure the description is short enough for autocomplete and `/prompts`
3. optionally mention the new mode in README docs

Because prompt selection is read from `SYSTEM_PROMPT`, changes here take effect immediately without schema migration.

## Slash Commands

Slash-command metadata and autocomplete live in [src/modules/glagentGui/commandSelector.go](C:/Users/amesa/Desktop/GlAgent/src/modules/glagentGui/commandSelector.go).

This file owns:

- command catalog
- provider suggestions
- model suggestions
- prompt suggestions
- computer-mode suggestions
- filtering logic
- selector rendering

If you add a new slash command and want autocomplete support, this is the file to update.

## Rendering

Two rendering paths exist:

- the main TUI uses Bubble Tea viewport rendering plus Glamour for assistant markdown
- `consoleMarkdown` is a separate terminal markdown utility used by the older streaming path

`consoleMarkdown` is still present and useful, but the TUI is now the main interface.

## Files Written by the App

GlAgent currently writes:

- `.env`
- `memory.json`
- `.glagent/sessions/*.json`

It may also read:

- local repo files
- provider environment variables
- Ollama API metadata

## Testing and Verification

Current verification is simple:

```bash
gofmt -w .
go test ./...
go build ./...
```

Today there are no deep automated tests for:

- command-loop behavior
- session restoration correctness
- prompt composition
- slash-command parsing

These are strong candidates for the next test pass.

## Recommended Next Engineering Improvements

### Safety

- replace string-block rules with structured policy checks
- add approval prompts for sensitive commands
- add cwd restrictions in workspace mode
- add audit logging for executed commands

### UX

- stream command output live into the viewport
- show executed commands in a dedicated UI panel
- improve error formatting for provider failures
- add session picker UI

### Architecture

- separate orchestration from UI state
- add structured tool-call schemas instead of tag parsing
- support non-Windows shells cleanly
- introduce integration tests around the full agent loop

## Extension Checklist

If you are making a non-trivial feature, check these areas:

- UI state in `glagentGui`
- prompt or context changes in `agentMod`
- persistence impact in `sessionstore`
- safety impact in `computer`
- discoverability in `README.md`
- developer notes in this file

## Known Constraints

- execution is PowerShell-specific today
- workspace safety is heuristic
- provider context is prompt-text based, not fully structured message objects
- command results are inserted as system messages, which is simple but not yet strongly typed

## Mental Model for Contributors

The cleanest way to think about GlAgent is:

- `glagentGui` is the conductor
- `agentMod` is the prompt + provider bridge
- `computer` is the supervised tool runner
- `memory` and `sessionstore` are persistence helpers
- `prompts` is behavior configuration

If you preserve that separation, the codebase stays easy to extend.

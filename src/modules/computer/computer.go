package computer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PermissionMode string

const (
	PermissionOff       PermissionMode = "off"
	PermissionWorkspace PermissionMode = "workspace"
	PermissionFull      PermissionMode = "full"
)

const (
	commandOpenTag  = "<glagent_command>"
	commandCloseTag = "</glagent_command>"
)

type CommandRequest struct {
	Command string
}

type ExecutionResult struct {
	Command    string
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
	TimedOut   bool
	WorkingDir string
}

func ParsePermissionMode(value string) PermissionMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(PermissionOff):
		return PermissionOff
	case string(PermissionWorkspace):
		return PermissionWorkspace
	case string(PermissionFull):
		return PermissionFull
	default:
		return ""
	}
}

func (m PermissionMode) String() string {
	if m == "" {
		return string(PermissionOff)
	}
	return string(m)
}

func (m PermissionMode) AllowsExecution() bool {
	return m == PermissionWorkspace || m == PermissionFull
}

func Instructions(mode PermissionMode) string {
	var b strings.Builder
	b.WriteString("You are GlAgent, a local AI agent inside a terminal app.\n")
	b.WriteString("If the user asks you to inspect the machine, run code, verify versions, list files, or gather local facts, prefer executing terminal commands instead of merely suggesting them.\n")
	b.WriteString("When you need the app to execute a command, respond with one or more blocks in exactly this format:\n")
	b.WriteString("<glagent_command>\n")
	b.WriteString("npm -v\n")
	b.WriteString("</glagent_command>\n")
	b.WriteString("After the command block(s), include a short plain-language note for the user.\n")
	b.WriteString("Do not wrap commands in markdown fences when using command blocks.\n")
	b.WriteString("Do not invent outputs. Wait for the command result before giving the final answer.\n")
	b.WriteString("Prefer concise, single-purpose commands. Avoid chaining commands unless necessary.\n")
	switch mode {
	case PermissionWorkspace:
		b.WriteString("Command execution is enabled in workspace mode. You may run development and inspection commands in the current project, but avoid destructive or machine-wide commands.\n")
	case PermissionFull:
		b.WriteString("Command execution is enabled in full-control mode. You may run broader system commands when needed, but use care and explain risky steps briefly.\n")
	default:
		b.WriteString("Command execution is disabled. Do not emit command blocks unless the user explicitly enables computer control.\n")
	}
	return b.String()
}

func ExtractCommands(response string) ([]CommandRequest, string) {
	var commands []CommandRequest
	var cleaned strings.Builder
	remaining := response

	for {
		start := strings.Index(remaining, commandOpenTag)
		if start == -1 {
			cleaned.WriteString(remaining)
			break
		}

		cleaned.WriteString(remaining[:start])
		remaining = remaining[start+len(commandOpenTag):]

		end := strings.Index(remaining, commandCloseTag)
		if end == -1 {
			cleaned.WriteString(commandOpenTag)
			cleaned.WriteString(remaining)
			break
		}

		command := strings.TrimSpace(remaining[:end])
		if command != "" {
			commands = append(commands, CommandRequest{Command: command})
		}
		remaining = remaining[end+len(commandCloseTag):]
	}

	return commands, strings.TrimSpace(cleaned.String())
}

func Execute(command string, mode PermissionMode, timeout time.Duration) (ExecutionResult, error) {
	if !mode.AllowsExecution() {
		return ExecutionResult{}, errors.New("computer control is disabled")
	}
	if mode == PermissionWorkspace && looksDangerous(command) {
		return ExecutionResult{}, fmt.Errorf("blocked potentially dangerous command in workspace mode: %s", command)
	}

	wd, err := os.Getwd()
	if err != nil {
		return ExecutionResult{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", command)
	cmd.Dir = wd

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	timedOut := ctx.Err() == context.DeadlineExceeded
	if runErr != nil {
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else if timedOut {
			exitCode = -1
		} else {
			return ExecutionResult{}, runErr
		}
	}

	return ExecutionResult{
		Command:    command,
		Stdout:     strings.TrimSpace(stdout.String()),
		Stderr:     strings.TrimSpace(stderr.String()),
		ExitCode:   exitCode,
		Duration:   duration,
		TimedOut:   timedOut,
		WorkingDir: wd,
	}, nil
}

func FormatResult(result ExecutionResult) string {
	var b strings.Builder
	b.WriteString("Command result:\n")
	b.WriteString("Command: ")
	b.WriteString(result.Command)
	b.WriteString("\n")
	b.WriteString("Working Directory: ")
	b.WriteString(result.WorkingDir)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Exit Code: %d\n", result.ExitCode))
	b.WriteString(fmt.Sprintf("Duration: %s\n", result.Duration.Round(time.Millisecond)))
	if result.TimedOut {
		b.WriteString("Timed Out: true\n")
	}
	if result.Stdout != "" {
		b.WriteString("Stdout:\n")
		b.WriteString(result.Stdout)
		b.WriteString("\n")
	}
	if result.Stderr != "" {
		b.WriteString("Stderr:\n")
		b.WriteString(result.Stderr)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func looksDangerous(command string) bool {
	lower := strings.ToLower(command)
	blocked := []string{
		"remove-item",
		"del ",
		"erase ",
		"rm ",
		"rmdir",
		"format ",
		"shutdown",
		"restart-computer",
		"stop-computer",
		"taskkill",
		"stop-process",
		"sc.exe delete",
	}

	for _, token := range blocked {
		if strings.Contains(lower, token) {
			return true
		}
	}

	return false
}

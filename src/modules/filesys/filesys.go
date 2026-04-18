package filesys

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"glagent/src/modules/computer"
)

const (
	readOpenTag   = "<glagent_file_read>"
	readCloseTag  = "</glagent_file_read>"
	listOpenTag   = "<glagent_file_list>"
	listCloseTag  = "</glagent_file_list>"
	writeTagStart = "<glagent_file_write"
	writeTagEnd   = "</glagent_file_write>"
	appendTagEnd  = "</glagent_file_append>"
	deleteTagEnd  = "</glagent_file_delete>"
	mkdirTagEnd   = "</glagent_file_mkdir>"
	moveTagEnd    = "</glagent_file_move>"
	patchTagEnd   = "</glagent_file_patch>"
)

type OperationType string

const (
	OpRead   OperationType = "read"
	OpList   OperationType = "list"
	OpWrite  OperationType = "write"
	OpAppend OperationType = "append"
	OpDelete OperationType = "delete"
	OpMkdir  OperationType = "mkdir"
	OpMove   OperationType = "move"
	OpPatch  OperationType = "patch"
)

type Request struct {
	Type       OperationType
	Path       string
	TargetPath string
	Content    string
	OldText    string
	NewText    string
}

type Result struct {
	Type       OperationType
	Path       string
	TargetPath string
	Content    string
	Entries    []string
	Bytes      int
	WorkingDir string
}

var (
	writeTagPattern  = regexp.MustCompile(`(?s)<glagent_file_write\s+path="([^"]+)">\s*(.*?)\s*</glagent_file_write>`)
	appendTagPattern = regexp.MustCompile(`(?s)<glagent_file_append\s+path="([^"]+)">\s*(.*?)\s*</glagent_file_append>`)
	deleteTagPattern = regexp.MustCompile(`(?s)<glagent_file_delete\s+path="([^"]+)">\s*</glagent_file_delete>`)
	mkdirTagPattern  = regexp.MustCompile(`(?s)<glagent_file_mkdir\s+path="([^"]+)">\s*</glagent_file_mkdir>`)
	moveTagPattern   = regexp.MustCompile(`(?s)<glagent_file_move\s+from="([^"]+)"\s+to="([^"]+)">\s*</glagent_file_move>`)
	patchTagPattern  = regexp.MustCompile(`(?s)<glagent_file_patch\s+path="([^"]+)">\s*<<OLD>>\s*(.*?)\s*<</OLD>>\s*<<NEW>>\s*(.*?)\s*<</NEW>>\s*</glagent_file_patch>`)
)

func Instructions(mode computer.PermissionMode) string {
	var b strings.Builder
	b.WriteString("You are operating with built-in local tools. Prefer built-in file actions over shell commands for repo file work.\n")
	b.WriteString("Before editing code, inspect the relevant files first. Prefer targeted edits over full-file rewrites.\n")
	b.WriteString("Use this workflow for coding tasks: understand -> inspect -> edit -> verify -> summarize.\n")
	b.WriteString("If a local fact can be verified from the machine or files, verify it before answering.\n")
	b.WriteString("To read a file, emit:\n")
	b.WriteString("<glagent_file_read>\nrelative/or/absolute/path.txt\n</glagent_file_read>\n")
	b.WriteString("To list a directory, emit:\n")
	b.WriteString("<glagent_file_list>\n.\n</glagent_file_list>\n")
	b.WriteString("To write or replace a file, emit:\n")
	b.WriteString("<glagent_file_write path=\"relative/or/absolute/path.txt\">\nnew file content here\n</glagent_file_write>\n")
	b.WriteString("To append to a file, emit:\n")
	b.WriteString("<glagent_file_append path=\"relative/or/absolute/path.txt\">\ncontent to append\n</glagent_file_append>\n")
	b.WriteString("To make a directory, emit:\n")
	b.WriteString("<glagent_file_mkdir path=\"relative/or/absolute/dir\"></glagent_file_mkdir>\n")
	b.WriteString("To move or rename a file, emit:\n")
	b.WriteString("<glagent_file_move from=\"old/path.txt\" to=\"new/path.txt\"></glagent_file_move>\n")
	b.WriteString("To delete a file or directory, emit:\n")
	b.WriteString("<glagent_file_delete path=\"path/to/delete\"></glagent_file_delete>\n")
	b.WriteString("To patch text inside a file, emit:\n")
	b.WriteString("<glagent_file_patch path=\"relative/or/absolute/path.txt\">\n<<OLD>>\nexact old text\n<</OLD>>\n<<NEW>>\nreplacement text\n<</NEW>>\n</glagent_file_patch>\n")
	b.WriteString("Use exact text for <<OLD>>. If the exact text is unknown, read the file first.\n")
	b.WriteString("Do not use markdown fences inside file blocks.\n")
	b.WriteString("Do not invent file contents. Wait for the app to return the real file result.\n")
	switch mode {
	case computer.PermissionWorkspace:
		b.WriteString("In workspace mode, file actions are limited to the current project directory.\n")
	case computer.PermissionFull:
		b.WriteString("In full-control mode, broader file access is allowed, but still prefer targeted file actions over shell commands.\n")
	default:
		b.WriteString("If computer control is off, do not emit file-action blocks.\n")
	}
	return b.String()
}

func ExtractRequests(response string) ([]Request, string) {
	var requests []Request
	working := response

	for {
		path, found, updated := extractSimpleBlock(working, readOpenTag, readCloseTag)
		if !found {
			break
		}
		if strings.TrimSpace(path) != "" {
			requests = append(requests, Request{Type: OpRead, Path: strings.TrimSpace(path)})
		}
		working = updated
	}

	for {
		path, found, updated := extractSimpleBlock(working, listOpenTag, listCloseTag)
		if !found {
			break
		}
		if strings.TrimSpace(path) != "" {
			requests = append(requests, Request{Type: OpList, Path: strings.TrimSpace(path)})
		}
		working = updated
	}

	working, requests = extractPatternRequests(working, writeTagPattern, requests, func(groups []string) Request {
		return Request{Type: OpWrite, Path: strings.TrimSpace(groups[0]), Content: strings.Trim(groups[1], "\r\n")}
	})
	working, requests = extractPatternRequests(working, appendTagPattern, requests, func(groups []string) Request {
		return Request{Type: OpAppend, Path: strings.TrimSpace(groups[0]), Content: strings.Trim(groups[1], "\r\n")}
	})
	working, requests = extractPatternRequests(working, deleteTagPattern, requests, func(groups []string) Request {
		return Request{Type: OpDelete, Path: strings.TrimSpace(groups[0])}
	})
	working, requests = extractPatternRequests(working, mkdirTagPattern, requests, func(groups []string) Request {
		return Request{Type: OpMkdir, Path: strings.TrimSpace(groups[0])}
	})
	working, requests = extractPatternRequests(working, moveTagPattern, requests, func(groups []string) Request {
		return Request{Type: OpMove, Path: strings.TrimSpace(groups[0]), TargetPath: strings.TrimSpace(groups[1])}
	})
	working, requests = extractPatternRequests(working, patchTagPattern, requests, func(groups []string) Request {
		return Request{
			Type:    OpPatch,
			Path:    strings.TrimSpace(groups[0]),
			OldText: normalizeBlock(groups[1]),
			NewText: normalizeBlock(groups[2]),
		}
	})

	return requests, strings.TrimSpace(working)
}

func Apply(request Request, mode computer.PermissionMode) (Result, error) {
	if !mode.AllowsExecution() {
		return Result{}, errors.New("computer control is disabled")
	}

	resolved, err := resolvePath(request.Path, mode)
	if err != nil {
		return Result{}, err
	}

	switch request.Type {
	case OpRead:
		data, err := os.ReadFile(resolved)
		if err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, Content: string(data), Bytes: len(data), WorkingDir: workspaceRoot()}, nil
	case OpList:
		entries, err := os.ReadDir(resolved)
		if err != nil {
			return Result{}, err
		}
		items := make([]string, 0, len(entries))
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			items = append(items, name)
		}
		sort.Strings(items)
		return Result{Type: request.Type, Path: resolved, Entries: items, WorkingDir: workspaceRoot()}, nil
	case OpWrite:
		if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
			return Result{}, err
		}
		data := []byte(request.Content)
		if err := os.WriteFile(resolved, data, 0644); err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, Bytes: len(data), WorkingDir: workspaceRoot()}, nil
	case OpAppend:
		if err := os.MkdirAll(filepath.Dir(resolved), 0755); err != nil {
			return Result{}, err
		}
		appendContent := request.Content
		if appendContent == "" {
			return Result{}, errors.New("append content cannot be empty")
		}
		f, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return Result{}, err
		}
		defer f.Close()
		n, err := f.WriteString(appendContent)
		if err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, Bytes: n, WorkingDir: workspaceRoot()}, nil
	case OpMkdir:
		if err := os.MkdirAll(resolved, 0755); err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, WorkingDir: workspaceRoot()}, nil
	case OpDelete:
		if err := os.RemoveAll(resolved); err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, WorkingDir: workspaceRoot()}, nil
	case OpMove:
		target, err := resolvePath(request.TargetPath, mode)
		if err != nil {
			return Result{}, err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return Result{}, err
		}
		if err := os.Rename(resolved, target); err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, TargetPath: target, WorkingDir: workspaceRoot()}, nil
	case OpPatch:
		data, err := os.ReadFile(resolved)
		if err != nil {
			return Result{}, err
		}
		current := string(data)
		if request.OldText == "" {
			return Result{}, errors.New("patch old text cannot be empty")
		}
		if !strings.Contains(current, request.OldText) {
			return Result{}, fmt.Errorf("patch target text not found in %s", request.Path)
		}
		updated := strings.Replace(current, request.OldText, request.NewText, 1)
		if err := os.WriteFile(resolved, []byte(updated), 0644); err != nil {
			return Result{}, err
		}
		return Result{Type: request.Type, Path: resolved, Bytes: len(updated), WorkingDir: workspaceRoot()}, nil
	default:
		return Result{}, fmt.Errorf("unsupported file request type: %s", request.Type)
	}
}

func AssessRisk(request Request, mode computer.PermissionMode) (bool, string) {
	switch request.Type {
	case OpDelete:
		return true, "delete requested"
	case OpMove:
		return request.Path != request.TargetPath, "move or rename requested"
	case OpWrite:
		return isSensitivePath(request.Path) || mode == computer.PermissionFull, "full file rewrite requested"
	case OpAppend:
		return isSensitivePath(request.Path), "append to sensitive file requested"
	case OpPatch:
		return strings.Contains(request.Path, ".gitignore") || isSensitivePath(request.Path), "patching a sensitive file requested"
	default:
		return false, ""
	}
}

func FormatResult(result Result) string {
	var b strings.Builder
	switch result.Type {
	case OpRead:
		b.WriteString("File read result:\n")
		b.WriteString("Path: ")
		b.WriteString(result.Path)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Bytes: %d\n", result.Bytes))
		b.WriteString("Content:\n")
		b.WriteString(result.Content)
	case OpList:
		b.WriteString("Directory listing result:\n")
		b.WriteString("Path: ")
		b.WriteString(result.Path)
		b.WriteString("\nEntries:\n")
		for _, entry := range result.Entries {
			b.WriteString("- ")
			b.WriteString(entry)
			b.WriteString("\n")
		}
	case OpWrite:
		b.WriteString("File write result:\nPath: ")
		b.WriteString(result.Path)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Bytes Written: %d", result.Bytes))
	case OpAppend:
		b.WriteString("File append result:\nPath: ")
		b.WriteString(result.Path)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Bytes Appended: %d", result.Bytes))
	case OpMkdir:
		b.WriteString("Directory create result:\nPath: ")
		b.WriteString(result.Path)
	case OpDelete:
		b.WriteString("Delete result:\nPath: ")
		b.WriteString(result.Path)
	case OpMove:
		b.WriteString("Move result:\nFrom: ")
		b.WriteString(result.Path)
		b.WriteString("\nTo: ")
		b.WriteString(result.TargetPath)
	case OpPatch:
		b.WriteString("Patch result:\nPath: ")
		b.WriteString(result.Path)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Updated Bytes: %d", result.Bytes))
	}
	return strings.TrimSpace(b.String())
}

func extractSimpleBlock(input, openTag, closeTag string) (string, bool, string) {
	start := strings.Index(input, openTag)
	if start == -1 {
		return "", false, input
	}
	end := strings.Index(input[start+len(openTag):], closeTag)
	if end == -1 {
		return "", false, input
	}
	contentStart := start + len(openTag)
	contentEnd := contentStart + end
	content := input[contentStart:contentEnd]
	updated := input[:start] + input[contentEnd+len(closeTag):]
	return content, true, updated
}

func extractPatternRequests(input string, pattern *regexp.Regexp, requests []Request, build func([]string) Request) (string, []Request) {
	matches := pattern.FindAllStringSubmatchIndex(input, -1)
	if len(matches) == 0 {
		return input, requests
	}

	var cleaned strings.Builder
	last := 0
	for _, match := range matches {
		cleaned.WriteString(input[last:match[0]])
		var groups []string
		for i := 2; i < len(match); i += 2 {
			groups = append(groups, input[match[i]:match[i+1]])
		}
		requests = append(requests, build(groups))
		last = match[1]
	}
	cleaned.WriteString(input[last:])
	return cleaned.String(), requests
}

func resolvePath(path string, mode computer.PermissionMode) (string, error) {
	root := workspaceRoot()
	if path == "" {
		return "", errors.New("empty path")
	}

	target := path
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, filepath.FromSlash(target))
	}

	resolved, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	if mode == computer.PermissionWorkspace {
		relative, err := filepath.Rel(root, resolved)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(relative, "..") {
			return "", fmt.Errorf("path %q is outside the workspace", path)
		}
	}

	return resolved, nil
}

func workspaceRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func isSensitivePath(path string) bool {
	lower := strings.ToLower(filepath.Base(path))
	return lower == ".env" || lower == ".gitignore" || lower == "go.mod" || lower == "go.sum"
}

func normalizeBlock(s string) string {
	return strings.Trim(s, "\r\n")
}

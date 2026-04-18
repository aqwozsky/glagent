package gitops

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Status() (string, error) {
	return runGit("status", "--short", "--branch")
}

func Diff(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return runGit("diff", "--", ".")
	}
	return runGit("diff", "--", path)
}

func Stage(path string) (string, error) {
	target := strings.TrimSpace(path)
	if target == "" {
		return "", fmt.Errorf("path is required")
	}
	if target == "." {
		if _, err := runGit("add", "."); err != nil {
			return "", err
		}
		return "Staged current workspace changes.", nil
	}
	if _, err := runGit("add", "--", target); err != nil {
		return "", err
	}
	return fmt.Sprintf("Staged %s.", target), nil
}

func Commit(message string) (string, error) {
	msg := strings.TrimSpace(message)
	if msg == "" {
		return "", fmt.Errorf("commit message is required")
	}
	if _, err := runGit("commit", "-m", msg); err != nil {
		return "", err
	}
	return fmt.Sprintf("Created commit: %s", msg), nil
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workspaceRoot()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			errText = err.Error()
		}
		return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), errText)
	}

	out := strings.TrimSpace(stdout.String())
	if out == "" {
		out = strings.TrimSpace(stderr.String())
	}
	return out, nil
}

func workspaceRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

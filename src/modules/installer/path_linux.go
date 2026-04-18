//go:build linux

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveInstallDir(scope Scope, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return filepath.Abs(override)
	}

	switch scope {
	case ScopeSystem:
		return "/usr/local/bin", nil
	case ScopeUser:
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", fmt.Errorf("could not determine user home directory")
		}
		return filepath.Join(home, ".local", "bin"), nil
	default:
		return "", fmt.Errorf("unsupported scope %q", scope)
	}
}

func ensurePathContains(scope Scope, installDir string) (bool, string, error) {
	if pathContains(os.Getenv("PATH"), installDir) {
		return false, "", nil
	}

	pathHint := fmt.Sprintf("Add %q to your PATH, for example: export PATH=%q:$PATH", installDir, installDir)
	return false, pathHint, nil
}

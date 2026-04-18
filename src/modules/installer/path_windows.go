//go:build windows

package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func resolveInstallDir(scope Scope, override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return filepath.Abs(override)
	}

	switch scope {
	case ScopeSystem:
		programFiles := os.Getenv("ProgramFiles")
		if programFiles == "" {
			return "", errors.New("ProgramFiles is not set")
		}
		return filepath.Join(programFiles, "GlAgent"), nil
	case ScopeUser:
		localAppData := os.Getenv("LocalAppData")
		if localAppData == "" {
			return "", errors.New("LocalAppData is not set")
		}
		return filepath.Join(localAppData, "Programs", "GlAgent"), nil
	default:
		return "", fmt.Errorf("unsupported scope %q", scope)
	}
}

func ensurePathContains(scope Scope, installDir string) (bool, string, error) {
	root, pathKey, err := pathRegistryTarget(scope)
	if err != nil {
		return false, "", err
	}

	key, err := registry.OpenKey(root, pathKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, "", err
	}
	defer key.Close()

	current, _, err := key.GetStringValue("Path")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return false, "", err
	}

	if pathContains(current, installDir) {
		return false, "", nil
	}

	updated := installDir
	if strings.TrimSpace(current) != "" {
		updated = current + ";" + installDir
	}

	if err := key.SetStringValue("Path", updated); err != nil {
		return false, "", err
	}
	return true, "", nil
}

func pathRegistryTarget(scope Scope) (registry.Key, string, error) {
	switch scope {
	case ScopeUser:
		return registry.CURRENT_USER, `Environment`, nil
	case ScopeSystem:
		return registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`, nil
	default:
		return 0, "", fmt.Errorf("unsupported scope %q", scope)
	}
}

//go:build !windows && !linux

package installer

import "fmt"

func resolveInstallDir(scope Scope, override string) (string, error) {
	return "", fmt.Errorf("glagent setup is not implemented for this operating system")
}

func ensurePathContains(scope Scope, installDir string) (bool, string, error) {
	return false, "", fmt.Errorf("glagent setup is not implemented for this operating system")
}

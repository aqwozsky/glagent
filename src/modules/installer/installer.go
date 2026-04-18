package installer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type Scope string

const (
	ScopeUser   Scope = "user"
	ScopeSystem Scope = "system"
)

type Options struct {
	Scope      Scope
	InstallDir string
	BinaryName string
}

type Result struct {
	Scope        Scope
	InstallDir   string
	BinaryPath   string
	PathUpdated  bool
	ResumeHint   string
	BuildCommand string
}

func Run(options Options) (Result, error) {
	if runtime.GOOS != "windows" {
		return Result{}, errors.New("glagent setup is currently implemented for Windows only")
	}

	scope := options.Scope
	if scope == "" {
		scope = ScopeUser
	}

	binaryName := options.BinaryName
	if strings.TrimSpace(binaryName) == "" {
		binaryName = "glagent.exe"
	}
	if !strings.HasSuffix(strings.ToLower(binaryName), ".exe") {
		binaryName += ".exe"
	}

	installDir, err := resolveInstallDir(scope, options.InstallDir)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return Result{}, err
	}

	tmpExe := filepath.Join(os.TempDir(), binaryName)
	buildCmd := fmt.Sprintf("go build -o %q .", tmpExe)
	if err := buildBinary(tmpExe); err != nil {
		return Result{}, err
	}
	defer os.Remove(tmpExe)

	targetExe := filepath.Join(installDir, binaryName)
	if err := copyFile(tmpExe, targetExe); err != nil {
		return Result{}, err
	}

	updated, err := ensurePathContains(scope, installDir)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Scope:        scope,
		InstallDir:   installDir,
		BinaryPath:   targetExe,
		PathUpdated:  updated,
		ResumeHint:   fmt.Sprintf("%s --continue <chat-id>", strings.TrimSuffix(binaryName, ".exe")),
		BuildCommand: buildCmd,
	}, nil
}

func buildBinary(output string) error {
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Dir = workspaceRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(out))
		if text == "" {
			text = err.Error()
		}
		return fmt.Errorf("build failed: %s", text)
	}
	return nil
}

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

func ensurePathContains(scope Scope, installDir string) (bool, error) {
	root, pathKey, err := pathRegistryTarget(scope)
	if err != nil {
		return false, err
	}

	key, err := registry.OpenKey(root, pathKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return false, err
	}
	defer key.Close()

	current, _, err := key.GetStringValue("Path")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return false, err
	}

	if pathContains(current, installDir) {
		return false, nil
	}

	updated := installDir
	if strings.TrimSpace(current) != "" {
		updated = current + ";" + installDir
	}

	if err := key.SetStringValue("Path", updated); err != nil {
		return false, err
	}
	return true, nil
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

func pathContains(pathValue, dir string) bool {
	want := strings.ToLower(filepath.Clean(dir))
	for _, part := range strings.Split(pathValue, ";") {
		if strings.ToLower(filepath.Clean(strings.TrimSpace(part))) == want {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func workspaceRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

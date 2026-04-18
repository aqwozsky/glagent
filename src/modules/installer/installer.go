package installer

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	Scope         Scope
	InstallDir    string
	BinaryPath    string
	PathUpdated   bool
	PathHint      string
	ResumeHint    string
	BuildCommand  string
	InstallSource string
}

func Run(options Options) (Result, error) {
	scope := options.Scope
	if scope == "" {
		scope = ScopeUser
	}

	binaryName := options.BinaryName
	if strings.TrimSpace(binaryName) == "" {
		binaryName = defaultBinaryName()
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(binaryName), ".exe") {
		binaryName += ".exe"
	}

	installDir, err := resolveInstallDir(scope, options.InstallDir)
	if err != nil {
		return Result{}, err
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return Result{}, err
	}

	stagedBinary, installSource, cleanup, buildCmd, err := stageInstallBinary(binaryName)
	if err != nil {
		return Result{}, err
	}
	defer cleanup()

	targetExe := filepath.Join(installDir, binaryName)
	currentExe, _ := os.Executable()
	if samePath(currentExe, targetExe) {
		installSource = "current executable already installed"
	} else if !samePath(stagedBinary, targetExe) {
		if err := copyFile(stagedBinary, targetExe); err != nil {
			return Result{}, err
		}
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetExe, 0755); err != nil {
			return Result{}, err
		}
	}

	updated, pathHint, err := ensurePathContains(scope, installDir)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Scope:         scope,
		InstallDir:    installDir,
		BinaryPath:    targetExe,
		PathUpdated:   updated,
		PathHint:      pathHint,
		ResumeHint:    fmt.Sprintf("%s --continue <chat-id>", strings.TrimSuffix(binaryName, filepath.Ext(binaryName))),
		BuildCommand:  buildCmd,
		InstallSource: installSource,
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

func stageInstallBinary(binaryName string) (string, string, func(), string, error) {
	currentExe, err := os.Executable()
	if err == nil && strings.TrimSpace(currentExe) != "" {
		stagePath := filepath.Join(os.TempDir(), "glagent-install-"+binaryName)
		if err := copyFile(currentExe, stagePath); err == nil {
			return stagePath, "current executable", func() { _ = os.Remove(stagePath) }, "", nil
		}
	}

	tmpExe := filepath.Join(os.TempDir(), "glagent-build-"+binaryName)
	buildCmd := fmt.Sprintf("go build -o %q .", tmpExe)
	if err := buildBinary(tmpExe); err != nil {
		return "", "", nil, buildCmd, err
	}
	return tmpExe, "fresh build from source", func() { _ = os.Remove(tmpExe) }, buildCmd, nil
}

func defaultBinaryName() string {
	if runtime.GOOS == "windows" {
		return "glagent.exe"
	}
	return "glagent"
}

func samePath(a, b string) bool {
	left, err := filepath.Abs(a)
	if err != nil {
		left = a
	}
	right, err := filepath.Abs(b)
	if err != nil {
		right = b
	}
	return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
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

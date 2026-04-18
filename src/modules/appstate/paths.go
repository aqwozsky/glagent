package appstate

import (
	"os"
	"path/filepath"
)

const appName = "GlAgent"

func BaseDir() string {
	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		return filepath.Join(configDir, appName)
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, "."+appName)
	}

	return "." + string(filepath.Separator) + appName
}

func EnsureBaseDir() (string, error) {
	base := BaseDir()
	if err := os.MkdirAll(base, 0755); err != nil {
		return "", err
	}
	return base, nil
}

func EnvFilePath() string {
	cwdEnv := ".env"
	if _, err := os.Stat(cwdEnv); err == nil {
		return cwdEnv
	}
	return filepath.Join(BaseDir(), ".env")
}

func MemoryFilePath() string {
	return filepath.Join(BaseDir(), "memory.json")
}

func SessionsDir() string {
	return filepath.Join(BaseDir(), "sessions")
}

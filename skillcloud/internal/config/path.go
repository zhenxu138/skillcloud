package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func HomeDir() (string, error) {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home, nil
		}
		if home := os.Getenv("HOME"); home != "" {
			return home, nil
		}
		return "", errors.New("home directory not found: USERPROFILE and HOME are unset")
	}

	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home, nil
	}
	return "", errors.New("home directory not found: HOME and USERPROFILE are unset")
}

func ExpandHome(path string) (string, error) {
	switch {
	case path == "~":
		home, err := HomeDir()
		if err != nil {
			return "", err
		}
		return home, nil
	case strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`):
		home, err := HomeDir()
		if err != nil {
			return "", err
		}
		rest := strings.TrimPrefix(strings.TrimPrefix(path, "~/"), `~\`)
		rest = strings.ReplaceAll(rest, `\`, string(filepath.Separator))
		return filepath.Join(home, filepath.FromSlash(rest)), nil
	default:
		return path, nil
	}
}

func DefaultConfigPath() (string, error) {
	return ExpandHome("~/.skillcloud/config.yaml")
}

func DefaultIndexPath() (string, error) {
	return ExpandHome("~/.skillcloud/index.json")
}

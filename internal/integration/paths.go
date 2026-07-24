package integration

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

func appConfigDir(app string) string {
	switch app {
	case "eza":
		if dir := os.Getenv("EZA_CONFIG_DIR"); dir != "" {
			return dir
		}
		return filepath.Join(configDir(), "eza")
	case "nvim":
		return filepath.Join(localConfigDir(), "nvim")
	case "vscode":
		return filepath.Join(userConfigDir(), "Code", "User")
	case "yazi":
		if runtime.GOOS == "windows" {
			return filepath.Join(configDir(), "yazi", "config")
		}
		return filepath.Join(configDir(), "yazi")
	case "zed":
		if runtime.GOOS == "windows" {
			return filepath.Join(configDir(), "Zed")
		}
		return filepath.Join(configDir(), "zed")
	case "rio":
		if dir := os.Getenv("RIO_CONFIG_HOME"); dir != "" {
			return dir
		}
		return filepath.Join(localConfigDir(), "rio")
	default:
		return filepath.Join(configDir(), app)
	}
}

// appConfigFile returns the path to a file inside an app's config directory.
func appConfigFile(app string, elems ...string) string {
	return filepath.Join(append([]string{appConfigDir(app)}, elems...)...)
}

// configDir returns the user's config root: $XDG_CONFIG_HOME or ~/.config on
// Unix, the platform config dir on Windows.
func configDir() string {
	if runtime.GOOS == "windows" {
		return userConfigDir()
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	return filepath.Join(homeDir(), ".config")
}

// localConfigDir returns %LOCALAPPDATA% on Windows and configDir elsewhere.
func localConfigDir() string {
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return dir
		}
	}
	return configDir()
}

func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		slog.Warn("user config dir not found", "error", err)
	}
	return dir
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("user home dir not found", "error", err)
	}
	return home
}

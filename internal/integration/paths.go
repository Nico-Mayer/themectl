package integration

import (
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

type appPath struct {
	base    func() string
	rel     []string
	darwin  []string
	windows []string
}

func (p appPath) dir() string {
	rel := p.rel
	switch runtime.GOOS {
	case "darwin":
		if p.darwin != nil {
			rel = p.darwin
		}
	case "windows":
		if p.windows != nil {
			rel = p.windows
		}
	}
	base := xdgConfigDir
	if p.base != nil {
		base = p.base
	}
	return filepath.Join(append([]string{base()}, rel...)...)
}

var appDirs = map[string]appPath{
	"ghostty": {rel: []string{"ghostty"}},
	"helix":   {rel: []string{"helix"}},
	"eza":     {base: ezaConfigDir, rel: []string{"eza"}},
	"nvim":    {base: localConfigDir, rel: []string{"nvim"}},
	"yazi":    {rel: []string{"yazi"}, windows: []string{"yazi", "config"}},
	"zed":     {rel: []string{"zed"}, windows: []string{"Zed"}},
	"vscode":  {base: userConfigDir, rel: []string{"Code", "User"}},
}

func appConfigDir(app string) string {
	return appDirs[app].dir()
}

func appConfigFile(app string, elems ...string) string {
	return filepath.Join(append([]string{appConfigDir(app)}, elems...)...)
}

func xdgConfigDir() string {
	if runtime.GOOS == "windows" {
		return userConfigDir()
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	return filepath.Join(homeDir(), ".config")
}

func localConfigDir() string {
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
			return dir
		}
	}
	return xdgConfigDir()
}

func userConfigDir() string {
	confDir, err := os.UserConfigDir()
	if err != nil {
		slog.Warn("user config home not set")
	}

	return confDir
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("user home not set")
	}

	return home
}

func ezaConfigDir() string {
	if dir := os.Getenv("EZA_CONFIG_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(homeDir(), ".config")
}

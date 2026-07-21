package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Nico-Mayer/themectl/internal/git"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type Zed struct {
	SettingsPath string
	Installer    ExtensionInstaller
}

type ExtensionInstaller interface {
	Ensure(ref ExtensionRef) error
}

type ExtensionRef struct {
	URL string
}

func (Zed) Name() string {
	return "zed"
}

func (z Zed) Apply(t theme.Resolved) error {
	spec := t.Zed
	if spec == nil || spec.Theme == "" {
		return fmt.Errorf("theme %s has no zed override", t.ID())
	}

	if z.Installer != nil {
		for _, url := range spec.Extensions {
			url = git.NormalizeURL(url)
			if err := z.Installer.Ensure(ExtensionRef{URL: url}); err != nil {
				return err
			}
		}
	}

	data, err := os.ReadFile(z.SettingsPath)
	if err != nil {
		return fmt.Errorf("read zed settings: %w", err)
	}

	updated, err := setZedString(string(data), "theme", spec.Theme)
	if err != nil {
		return err
	}

	if spec.IconTheme == "" {
		spec.IconTheme = "Zed (Default)"
	}

	updated, err = setZedString(updated, "icon_theme", spec.IconTheme)
	if err != nil {
		return err
	}

	if err := os.WriteFile(z.SettingsPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write zed settings: %w", err)
	}

	return nil
}

func (z Zed) Check() error {
	return checkConfigDir(z.Name(), z.SettingsPath)
}

func setZedString(config, key, value string) (string, error) {
	quoted, _ := json.Marshal(value) // marshaling a string never fails

	re := regexp.MustCompile(`("` + regexp.QuoteMeta(key) + `"\s*:\s*)"[^"]*"`)
	if re.MatchString(config) {
		repl := `${1}` + strings.ReplaceAll(string(quoted), "$", "$$")
		return re.ReplaceAllString(config, repl), nil
	}

	end := strings.LastIndex(config, "}")
	if end < 0 {
		return "", fmt.Errorf("no object found in zed config")
	}
	head := strings.TrimRight(config[:end], " \t\r\n")
	if head == "" {
		return "", fmt.Errorf("no object found in zed config")
	}

	sep := ",\n"
	if last := head[len(head)-1]; last == '{' || last == ',' {
		sep = "\n"
	}
	return head + sep + "  \"" + key + "\": " + string(quoted) + "\n" + config[end:], nil
}

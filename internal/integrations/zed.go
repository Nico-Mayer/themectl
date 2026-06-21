package integrations

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/log"
	"github.com/go-git/go-billy/v5"
	billyiofs "github.com/go-git/go-billy/v5/helper/iofs"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/model"
	"github.com/nico-mayer/themectl-cli/internal/receipt"
)

type Zed struct{}

type ZedReceipt struct {
	Id           string `json:"id"`
	Version      string `json:"version"`
	ExtensionUrl string `json:"extension-url"`
}

type ZedThemeInfo struct {
	ExtensionUrl string `json:"extension-url"`
	Theme        string `json:"theme"`
}

type ExtensionManifest struct {
	ID          string   `toml:"id"`
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	SchemaVer   int      `toml:"schema_version"`
	Authors     []string `toml:"authors"`
	Description string   `toml:"description"`
	Repository  string   `toml:"repository"`
}

func init() {
	Register(Zed{})
}

func (Zed) Name() string {
	return "zed"
}

func (i Zed) Apply(themeInfo model.ThemeInfo) error {
	logger := integrationLogger(i)

	zedThemeInfo, err := loadZedThemeInfo()
	if err != nil {
		return err
	}

	hasReceipt := true
	_, err = receipt.Load[ZedReceipt](i.Name(), themeInfo.Name)
	if err != nil {
		log.Warn(err)
		hasReceipt = false
	}

	if !hasReceipt {
		logger.Debug("download extension", "url", zedThemeInfo.ExtensionUrl)
		extManifest, err := i.ensureExtension(zedThemeInfo)
		if err != nil {
			return err
		}

		receipt.Save(i.Name(), themeInfo.Name, ZedReceipt{
			Id:           extManifest.ID,
			Version:      extManifest.Version,
			ExtensionUrl: extManifest.Repository,
		})
	}

	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	zedSettingsDir := cfg.Settings.ConfigDirFor(i.Name())
	zedSettingsPath := filepath.Join(zedSettingsDir, "settings.json")

	data, err := os.ReadFile(zedSettingsPath)
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}

	re := regexp.MustCompile(`("theme"\s*:\s*")([^"]*)(")`)
	if !re.MatchString(string(data)) {
		return fmt.Errorf("no \"theme\" key found in %s", zedSettingsPath)
	}

	updated := re.ReplaceAllString(string(data), `${1}`+zedThemeInfo.Theme+`${3}`)
	if err := os.WriteFile(zedSettingsPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}

	logger.Info("theme applied", "theme", zedThemeInfo.Theme)
	return nil
}

func loadZedThemeInfo() (ZedThemeInfo, error) {
	cfg, _ := config.Get()
	data, err := os.ReadFile(filepath.Join(cfg.Paths.CurrentThemeDir, "zed.json"))
	if err != nil {
		return ZedThemeInfo{}, fmt.Errorf("read zed theme info: %w", err)
	}

	var info ZedThemeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return ZedThemeInfo{}, fmt.Errorf("parse zed theme info: %w", err)
	}
	return info, nil
}

func (i Zed) ensureExtension(info ZedThemeInfo) (ExtensionManifest, error) {
	logger := integrationLogger(i)

	r, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:   fmt.Sprintf("https://%s", info.ExtensionUrl),
		Depth: 1,
	})
	if err != nil {
		return ExtensionManifest{}, fmt.Errorf("clone extension: %w", err)
	}

	wt, err := r.Worktree()
	if err != nil {
		return ExtensionManifest{}, fmt.Errorf("read worktree: %w", err)
	}

	stdFs := billyiofs.New(wt.Filesystem)

	manifest, err := parseManifest(stdFs)
	if err != nil {
		return ExtensionManifest{}, err
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return ExtensionManifest{}, fmt.Errorf("resolve config dir: %w", err)
	}

	if runtime.GOOS == "windows" {
		userConfigDir = filepath.Join(filepath.Dir(userConfigDir), "Local")
	}

	targetDir := filepath.Join(userConfigDir, "Zed", "extensions", "installed", manifest.ID)

	if _, err := os.Stat(targetDir); err == nil {
		logger.Debug("already installed", "extension", manifest.ID)
		logger.Debug("clear", "dir", targetDir)
		err = os.RemoveAll(targetDir)
		if err != nil {
			logger.Error("deleting", "dir", targetDir, "err", err)
		}
	}

	logger.Debug("installing extension", "extension", manifest.ID, "target", targetDir)

	if err := copyToDir(wt.Filesystem, targetDir); err != nil {
		return ExtensionManifest{}, fmt.Errorf("install extension: %w", err)
	}

	logger.Info("extension installed", "extension", manifest.ID)
	return manifest, nil
}

func parseManifest(fsys fs.FS) (ExtensionManifest, error) {
	f, err := fsys.Open("extension.toml")
	if err != nil {
		return ExtensionManifest{}, fmt.Errorf("open extension.toml: %w", err)
	}
	defer f.Close()

	var manifest ExtensionManifest
	if _, err := toml.NewDecoder(f).Decode(&manifest); err != nil {
		return ExtensionManifest{}, fmt.Errorf("parse extension.toml: %w", err)
	}
	return manifest, nil
}

func copyToDir(srcFs billy.Filesystem, targetDir string) error {
	return util.Walk(srcFs, ".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dst := filepath.Join(targetDir, path)

		if info.IsDir() {
			return os.MkdirAll(dst, 0o755)
		}

		src, err := srcFs.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, src)
		return err
	})
}

package cli

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/store"
	"github.com/Nico-Mayer/themectl/internal/wallpaper"
)

func TestReadCurrentTheme_MissingIncludesRecovery(t *testing.T) {
	_, err := readCurrentTheme(filepath.Join(t.TempDir(), ".current"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("error = %v, want wrapped os.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "read current theme") || !strings.Contains(err.Error(), "`themectl set`") {
		t.Errorf("error %q does not identify operation and recovery command", err)
	}
}

func TestInvalidInputErrors_StateRequirement(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "missing install URL", args: []string{"themectl", "install"}, want: []string{"git repository URL is required"}},
		{name: "conflicting appearance flags", args: []string{"themectl", "list", "--light", "--dark"}, want: []string{"--light", "--dark", "cannot be used together"}},
		{name: "missing wallpaper path", args: []string{"themectl", "wallpaper", "set"}, want: []string{"wallpaper path is required", "--random"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testApp(t).Run(context.Background(), tt.args)
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, text := range tt.want {
				if !strings.Contains(err.Error(), text) {
					t.Errorf("error %q missing %q", err, text)
				}
			}
		})
	}
}

func TestPickWallpaper_NoCandidatesIncludesRecovery(t *testing.T) {
	_, err := pickWallpaper(nil)
	if err == nil || !strings.Contains(err.Error(), "`themectl wallpaper set <filepath>`") {
		t.Errorf("error = %v, want wallpaper file recovery command", err)
	}
}

func TestApplyRandomWallpaper_NoCandidatesReportsSkip(t *testing.T) {
	root := t.TempDir()
	cfg := config.Config{Root: root, CacheRoot: t.TempDir()}
	if err := store.WriteCurrent(cfg.CurrentFile(), "catppuccin/mocha"); err != nil {
		t.Fatal(err)
	}
	st := store.NewStore(testThemeFS(), nil)
	manager := wallpaper.NewManager(cfg.ThemesDir(), cfg.SharedWallpapersDir())

	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	if err := applyRandomWallpaper(cfg, st, manager); err != nil {
		t.Fatalf("applyRandomWallpaper returned error for preserved skip behavior: %v", err)
	}
	for _, text := range []string{"wallpaper not changed", "no candidates", "`themectl wallpaper set <filepath>`"} {
		if !strings.Contains(logs.String(), text) {
			t.Errorf("log %q missing %q", logs.String(), text)
		}
	}
	if strings.Contains(logs.String(), `msg="wallpaper set"`) {
		t.Errorf("skip was reported as success: %s", logs.String())
	}
}

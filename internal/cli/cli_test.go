package cli

import (
	"context"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/store"
	urfavecli "github.com/urfave/cli/v3"
)

func testApp(t *testing.T) *urfavecli.Command {
	t.Helper()
	cfg := config.Config{Root: t.TempDir(), CacheRoot: t.TempDir()}
	st := store.NewStore(testThemeFS(), nil)
	return New(cfg, st, nil)
}

func testThemeFS() fstest.MapFS {
	return fstest.MapFS{
		"catppuccin/theme.toml": {Data: []byte(`
[defaults]
appearance = "dark"

[variants.mocha]
wallpaper_sources = ["nature"]
`)},
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = old })

	runErr := fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old

	return string(out), runErr
}

func TestNew_CommandContract(t *testing.T) {
	tests := []struct {
		path    string
		aliases []string
		flags   []string
	}{
		{path: "list", aliases: []string{"ls"}, flags: []string{"dark", "json", "light"}},
		{path: "set", aliases: []string{"apply", "use"}},
		{path: "set random", flags: []string{"dark", "light"}},
		{path: "current", flags: []string{"json"}},
		{path: "wallpaper", aliases: []string{"wall"}},
		{path: "wallpaper list", aliases: []string{"ls"}},
		{path: "wallpaper set", flags: []string{"random"}},
		{path: "refresh", aliases: []string{"reapply"}},
		{path: "doctor", aliases: []string{"status"}, flags: []string{"json"}},
		{path: "install", flags: []string{"force", "name"}},
		{path: "uninstall"},
		{path: "update"},
		{path: "cache"},
		{path: "cache clear"},
	}

	app := testApp(t)
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			cmd := commandAt(t, app, tt.path)

			aliases := append([]string(nil), cmd.Aliases...)
			sort.Strings(aliases)
			sort.Strings(tt.aliases)
			if !reflect.DeepEqual(aliases, tt.aliases) {
				t.Fatalf("aliases = %v, want %v", aliases, tt.aliases)
			}

			flags := make([]string, 0, len(cmd.Flags))
			for _, flag := range cmd.Flags {
				names := flag.Names()
				flags = append(flags, names[0])
			}
			sort.Strings(flags)
			sort.Strings(tt.flags)
			if strings.Join(flags, ",") != strings.Join(tt.flags, ",") {
				t.Fatalf("flags = %v, want %v", flags, tt.flags)
			}
		})
	}
}

func TestNew_ArgumentAndFlagAliasContract(t *testing.T) {
	app := testApp(t)
	argumentTests := map[string]string{
		"set":            "theme",
		"wallpaper list": "theme",
		"wallpaper set":  "path",
		"install":        "url",
		"uninstall":      "family",
	}
	for path, want := range argumentTests {
		t.Run("argument/"+path, func(t *testing.T) {
			arg, ok := commandAt(t, app, path).Arguments[0].(*urfavecli.StringArg)
			if !ok {
				t.Fatalf("argument is %T, want *cli.StringArg", commandAt(t, app, path).Arguments[0])
			}
			if arg.Name != want {
				t.Errorf("argument name = %q, want %q", arg.Name, want)
			}
		})
	}

	flagTests := []struct {
		path string
		name string
		want []string
	}{
		{path: "list", name: "light", want: []string{"light", "l"}},
		{path: "list", name: "dark", want: []string{"dark", "d"}},
		{path: "install", name: "force", want: []string{"force", "f"}},
		{path: "wallpaper set", name: "random", want: []string{"random", "r"}},
	}
	for _, tt := range flagTests {
		t.Run("flag/"+tt.path+"/"+tt.name, func(t *testing.T) {
			got := flagAt(t, commandAt(t, app, tt.path), tt.name).Names()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flag names = %v, want %v", got, tt.want)
			}
		})
	}
}

func flagAt(t *testing.T, command *urfavecli.Command, name string) urfavecli.Flag {
	t.Helper()
	for _, flag := range command.Flags {
		if flag.Names()[0] == name {
			return flag
		}
	}
	t.Fatalf("flag %q not found on command %q", name, command.Name)
	return nil
}

func TestNew_GlobalFlagContract(t *testing.T) {
	app := testApp(t)
	if len(app.Flags) != 1 || !reflect.DeepEqual(app.Flags[0].Names(), []string{"verbose", "v"}) {
		t.Fatalf("global flags = %v, want verbose with v alias", app.Flags)
	}
}

func TestNew_MissingInstallArgumentReturnsError(t *testing.T) {
	err := testApp(t).Run(context.Background(), []string{"themectl", "install"})
	if err == nil {
		t.Fatal("expected missing install argument to return an error")
	}
}

func commandAt(t *testing.T, root *urfavecli.Command, path string) *urfavecli.Command {
	t.Helper()
	cmd := root
	for _, name := range strings.Fields(path) {
		var found *urfavecli.Command
		for _, child := range cmd.Commands {
			if child.Name == name {
				found = child
				break
			}
		}
		if found == nil {
			t.Fatalf("command %q not found", path)
		}
		cmd = found
	}
	return cmd
}

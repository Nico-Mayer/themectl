package theme

import (
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestResolve_variantOverridesFamily(t *testing.T) {
	fam := Family{
		Name: "catppuccin",
		Defaults: Spec{
			Appearance: new(Dark),
			Ghostty:    &GhosttySpec{Theme: "cat-default"},
			Zed:        &ZedSpec{Theme: "Cat Mocha", Extensions: []string{"github.com/catppuccin/zed"}},
		},
	}
	v := Variant{
		Name: "latte",
		Spec: Spec{
			Appearance:       new(Light),
			Ghostty:          &GhosttySpec{Theme: "catppuccin-latte"},
			WallpaperSources: []string{"catppuccin/macchiato"},
		},
	}

	got, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Equal(t, got.Appearance, Light)
	testutil.Equal(t, got.ID(), "catppuccin/latte")
	testutil.Diff(t, []string{"catppuccin/macchiato", "catppuccin/latte"}, got.WallpaperSources)
	testutil.Diff(t, map[string]string{"ghostty": "catppuccin-latte", "zed": "Cat Mocha"}, got.Themes())
}

func TestResolve_variantInheritsAppearance(t *testing.T) {
	fam := Family{Name: "f", Defaults: Spec{Appearance: new(Dark)}}
	v := Variant{Name: "v"}

	got, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Equal(t, got.Appearance, Dark)
}

func TestResolve_wallpaperSourcesIncludeOwnID(t *testing.T) {
	fam := Family{Name: "f", Defaults: Spec{Appearance: new(Dark)}}
	v := Variant{Name: "v"}

	got, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"f/v"}, got.WallpaperSources)
}

func TestResolve_wallpaperSourcesInheritedFromFamily(t *testing.T) {
	fam := Family{Name: "f", Defaults: Spec{Appearance: new(Dark), WallpaperSources: []string{"shared"}}}
	v := Variant{Name: "v"}

	got, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"shared", "f/v"}, got.WallpaperSources)
}

func TestResolve_wallpaperSourcesVariantOverridesFamily(t *testing.T) {
	fam := Family{Name: "f", Defaults: Spec{Appearance: new(Dark), WallpaperSources: []string{"shared"}}}
	v := Variant{Name: "v", Spec: Spec{WallpaperSources: []string{"own"}}}

	got, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Diff(t, []string{"own", "f/v"}, got.WallpaperSources)
}

func TestResolve_missingAppearanceFails(t *testing.T) {
	if _, err := Resolve(Family{Name: "f"}, Variant{Name: "v"}); err == nil {
		t.Fatal("expected error when appearance is set by neither family nor variant")
	}
}

func TestResolve_doesNotMutateInputs(t *testing.T) {
	fam := Family{Name: "f", Defaults: Spec{Appearance: new(Dark), Zed: &ZedSpec{Theme: "a", IconTheme: "a-icons"}}}
	v := Variant{Name: "v", Spec: Spec{Zed: &ZedSpec{Theme: "b"}}}

	_, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Equal(t, fam.Defaults.Zed.Theme, "a")
	testutil.Equal(t, v.Zed.IconTheme, "")
}

func TestResolve_vscode(t *testing.T) {
	appearance := Dark
	fam := Family{Name: "cat", Defaults: Spec{
		Appearance: &appearance,
		VSCode:     &VSCodeSpec{Theme: "Base", Extensions: []string{"catppuccin.catppuccin-vsc"}},
	}}
	v := Variant{Name: "mocha", Spec: Spec{
		VSCode: &VSCodeSpec{Theme: "Catppuccin Mocha"},
	}}

	res, err := Resolve(fam, v)
	testutil.NoErr(t, err)

	testutil.Equal(t, res.VSCode.Theme, "Catppuccin Mocha")
	testutil.Diff(t, []string{"catppuccin.catppuccin-vsc"}, res.VSCode.Extensions)
	testutil.Equal(t, res.Themes()["vscode"], "Catppuccin Mocha")
}

func TestResolve_assetTarget(t *testing.T) {
	fam := Family{
		Name: "cat",
		Defaults: Spec{
			Appearance: new(Dark),
			Nvim: &SymlinkSpec{
				URL: "https://test.com",
			},
			Eza: &SymlinkSpec{
				URL: "https://eza.com/theme",
			},
		},
	}

	v := Variant{
		Name: "mocha",
		Spec: Spec{
			Nvim: &SymlinkSpec{
				URL: "https://catppuccin/mocha/nvim",
			},
		},
	}

	res, err := Resolve(fam, v)
	testutil.NoErr(t, err)
	testutil.Equal(t, res.Nvim.URL, "https://catppuccin/mocha/nvim")
	testutil.Equal(t, res.Eza.URL, "https://eza.com/theme")

	got := res.RemoteAssets()
	testutil.Equal(t, len(got), 2)
	testutil.Equal(t, got[NvimAssetName], "https://catppuccin/mocha/nvim")
	testutil.Equal(t, got[EzaAssetName], "https://eza.com/theme")
}

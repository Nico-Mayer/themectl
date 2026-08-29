package integration

import (
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

// A settings file shaped like Windows Terminal's: the target key also appears
// inside profiles.list, where the old regex helper would have bound to it.
const wtSettings = `{
  // the default profile
  "defaultProfile": "{guid-a}",
  "profiles": {
    "defaults": {
      "font": { "face": "Cascadia Code" }
    },
    "list": [
      {
        "guid": "{guid-a}",
        "colorScheme": "Campbell"
      },
    ]
  },
  "schemes": [],
}`

func TestSetJSONPathNested(t *testing.T) {
	got, err := setJSONPath(wtSettings, []string{"profiles", "defaults", "colorScheme"}, "themectl")
	testutil.NoErr(t, err)

	want := `{
  // the default profile
  "defaultProfile": "{guid-a}",
  "profiles": {
    "defaults": {
      "font": { "face": "Cascadia Code" },
      "colorScheme": "themectl"
    },
    "list": [
      {
        "guid": "{guid-a}",
        "colorScheme": "Campbell"
      },
    ]
  },
  "schemes": [],
}`
	testutil.Equal(t, got, want)
}

func TestSetJSONPathCreatesMissingObjects(t *testing.T) {
	got, err := setJSONPath(`{"a": 1}`, []string{"profiles", "defaults", "colorScheme"}, "themectl")
	testutil.NoErr(t, err)
	testutil.Equal(t, got, "{\"a\": 1,\n  \"profiles\": {\n  \"defaults\": {\n  \"colorScheme\": \"themectl\"\n}\n}}")
}

func TestSetJSONPathErrors(t *testing.T) {
	tests := []struct {
		name string
		in   string
		path []string
	}{
		{name: "no object", in: `// just a comment`, path: []string{"theme"}},
		{name: "array root", in: `[1, 2]`, path: []string{"theme"}},
		{name: "intermediate is not an object", in: `{"profiles": "nope"}`, path: []string{"profiles", "defaults"}},
		{name: "invalid json", in: `{`, path: []string{"theme"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := setJSONPath(tt.in, tt.path, "x")
			testutil.Equal(t, err != nil, true)
		})
	}
}

func TestDeleteJSONPath(t *testing.T) {
	set, err := setJSONPath(wtSettings, []string{"profiles", "defaults", "colorScheme"}, "themectl")
	testutil.NoErr(t, err)

	got, err := deleteJSONPath(set, []string{"profiles", "defaults", "colorScheme"})
	testutil.NoErr(t, err)
	testutil.Equal(t, got, wtSettings)
}

func TestDeleteJSONPathMissingIsNoOp(t *testing.T) {
	tests := []struct {
		name string
		path []string
	}{
		{name: "missing leaf", path: []string{"profiles", "defaults", "colorScheme"}},
		{name: "missing branch", path: []string{"nope", "colorScheme"}},
		{name: "branch is not an object", path: []string{"defaultProfile", "colorScheme"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := deleteJSONPath(wtSettings, tt.path)
			testutil.NoErr(t, err)
			testutil.Equal(t, got, wtSettings)
		})
	}
}

const themectlChrome = `{
  "name": "themectl",
  "window": { "applicationTheme": "dark" }
}`

func TestUpsertJSONArrayByNameAppends(t *testing.T) {
	in := `{
  "themes": [
    {
      "name": "mine",
      "window": { "applicationTheme": "light" }
    }
  ]
}`
	got, err := upsertJSONArrayByName(in, []string{"themes"}, "themectl", []byte(themectlChrome))
	testutil.NoErr(t, err)

	want := `{
  "themes": [
    {
      "name": "mine",
      "window": { "applicationTheme": "light" }
    },
    {
  "name": "themectl",
  "window": { "applicationTheme": "dark" }
}
  ]
}`
	testutil.Equal(t, got, want)
}

// The upsert is on the hot path, so applying repeatedly must converge.
func TestUpsertJSONArrayByNameIsIdempotent(t *testing.T) {
	in := `{"themes": [{"name": "mine"}]}`

	once, err := upsertJSONArrayByName(in, []string{"themes"}, "themectl", []byte(`{"name":"themectl","v":1}`))
	testutil.NoErr(t, err)
	twice, err := upsertJSONArrayByName(once, []string{"themes"}, "themectl", []byte(`{"name":"themectl","v":1}`))
	testutil.NoErr(t, err)
	testutil.Equal(t, twice, once)

	replaced, err := upsertJSONArrayByName(once, []string{"themes"}, "themectl", []byte(`{"name":"themectl","v":2}`))
	testutil.NoErr(t, err)
	testutil.Equal(t, replaced, `{"themes": [{"name": "mine"},{"name":"themectl","v":2}]}`)
}

func TestUpsertJSONArrayByNameCreatesArray(t *testing.T) {
	got, err := upsertJSONArrayByName(`{}`, []string{"themes"}, "themectl", []byte(`{"name":"themectl"}`))
	testutil.NoErr(t, err)
	testutil.Equal(t, got, "{\n  \"themes\": [{\"name\":\"themectl\"}]\n}")
}

func TestRemoveJSONArrayByName(t *testing.T) {
	in := `{"themes": [{"name": "mine"}, {"name": "themectl"}, {"name": "other"}]}`

	got, err := removeJSONArrayByName(in, []string{"themes"}, "themectl")
	testutil.NoErr(t, err)
	testutil.Equal(t, got, `{"themes": [{"name": "mine"}, {"name": "other"}]}`)
}

func TestRemoveJSONArrayByNameMissingIsNoOp(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "no such element", in: `{"themes": [{"name": "mine"}]}`},
		{name: "no array", in: `{"a": 1}`},
		{name: "not an array", in: `{"themes": "nope"}`},
		{name: "elements without a name", in: `{"themes": [1, "two", {"other": "key"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeJSONArrayByName(tt.in, []string{"themes"}, "themectl")
			testutil.NoErr(t, err)
			testutil.Equal(t, got, tt.in)
		})
	}
}

// Windows Terminal ships settings.json with comments, so an edit that strips
// them is a regression the user sees. Apply then reset must also be a round
// trip: nothing themectl added stays behind.
func TestHujsonPreservesUntouchedBytes(t *testing.T) {
	const messy = "// leading comment with a { brace\n" +
		"{\n" +
		"\t/* block */ \"defaultProfile\":   \"{guid-a}\",\n" +
		"\t\"themes\": [\n" +
		"\t\t{ \"name\": \"mine\" },\n" +
		"\t],\n" +
		"\t\"profiles\": {\n" +
		"\t\t\"defaults\": { \"font\": \"Cascadia\" },\n" +
		"\t\t\"list\": [ { \"colorScheme\": \"Campbell\" } ],\n" +
		"\t}, // trailing note\n" +
		"}\n"

	apply := []struct {
		name string
		fn   func(string) (string, error)
	}{
		{"set colorScheme", func(s string) (string, error) {
			return setJSONPath(s, []string{"profiles", "defaults", "colorScheme"}, "themectl")
		}},
		{"upsert chrome", func(s string) (string, error) {
			return upsertJSONArrayByName(s, []string{"themes"}, "themectl", []byte(`{"name":"themectl"}`))
		}},
	}
	reset := []struct {
		name string
		fn   func(string) (string, error)
	}{
		{"remove chrome", func(s string) (string, error) {
			return removeJSONArrayByName(s, []string{"themes"}, "themectl")
		}},
		{"delete colorScheme", func(s string) (string, error) {
			return deleteJSONPath(s, []string{"profiles", "defaults", "colorScheme"})
		}},
	}

	got := messy
	var err error
	for _, step := range append(apply, reset...) {
		got, err = step.fn(got)
		testutil.NoErr(t, err)
		for _, keep := range []string{
			"// leading comment with a { brace",
			"/* block */",
			"// trailing note",
			"\t\t{ \"name\": \"mine\" },",
			"\"colorScheme\": \"Campbell\"",
		} {
			if !strings.Contains(got, keep) {
				t.Fatalf("after %s, %q was lost:\n%s", step.name, keep, got)
			}
		}
	}

	testutil.Equal(t, got, messy)
}

// Ported from the regex helper this replaces, so the swap is proven against
// the behavior it supersedes. Formatting of an appended key differs; the
// semantics do not.
func TestSetJSONPathPortedFromJSONC(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		key     string
		value   string
		want    string
		wantErr bool
	}{
		{name: "simple", in: `{"theme": "old"}`, key: "theme", value: "One Dark", want: `{"theme": "One Dark"}`},
		{name: "loose spacing", in: `{"theme"   :   "old"}`, key: "theme", value: "new", want: `{"theme"   :   "new"}`},
		{
			name:  "preserves siblings and comments",
			in:    "{\n  // pick a theme\n  \"theme\": \"old\",\n  \"vim_mode\": true\n}",
			key:   "theme",
			value: "Catppuccin Mocha",
			want:  "{\n  // pick a theme\n  \"theme\": \"Catppuccin Mocha\",\n  \"vim_mode\": true\n}",
		},
		{name: "icon_theme", in: `{"icon_theme": "old"}`, key: "icon_theme", value: "new", want: `{"icon_theme": "new"}`},
		{
			name: "missing key appended", in: `{"vim_mode": true}`, key: "theme", value: "new",
			want: "{\"vim_mode\": true,\n  \"theme\": \"new\"}",
		},
		{
			name: "missing key appended after trailing comma", in: "{\n  \"vim_mode\": true,\n}", key: "theme", value: "new",
			want: "{\n  \"vim_mode\": true,\n  \"theme\": \"new\",\n}",
		},
		{
			name: "missing key appended to empty object", in: `{}`, key: "theme", value: "new",
			want: "{\n  \"theme\": \"new\"\n}",
		},
		{
			name: "value with quotes escaped", in: `{}`, key: "theme", value: `say "hi"`,
			want: "{\n  \"theme\": \"say \\\"hi\\\"\"\n}",
		},
		{
			name: "replaced value with dollar sign", in: `{"theme": "old"}`, key: "theme", value: "a$1b",
			want: `{"theme": "a$1b"}`,
		},
		{
			name: "leading comment with brace ignored", in: "// {settings}\n{\n  \"vim_mode\": true\n}", key: "theme", value: "new",
			want: "// {settings}\n{\n  \"vim_mode\": true,\n  \"theme\": \"new\"\n}",
		},
		{name: "no object", in: `// just a comment`, key: "theme", value: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setJSONPath(tt.in, []string{tt.key}, tt.value)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
}

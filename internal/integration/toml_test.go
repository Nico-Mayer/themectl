package integration

import (
	"testing"

	"github.com/Nico-Mayer/themectl/internal/testutil"
)

func TestSetTOMLString(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		key     string
		value   string
		want    string
		wantErr bool
	}{
		{name: "simple", in: `theme = "old"`, key: "theme", value: "new", want: `theme = "new"`},
		{name: "loose spacing", in: `theme   =   "old"`, key: "theme", value: "new", want: `theme   =   "new"`},
		{name: "unquoted value replaced", in: `theme = old`, key: "theme", value: "new", want: `theme = "new"`},
		{name: "preserves siblings",
			in:    "theme = \"old\"\n[editor]\nline-number = \"relative\"\n",
			key:   "theme",
			value: "catppuccin_mocha",
			want:  "theme = \"catppuccin_mocha\"\n[editor]\nline-number = \"relative\"\n"},
		{name: "key must start the line",
			in:    "icon-theme = \"old\"\ntheme = \"old\"\n",
			key:   "theme",
			value: "new",
			want:  "icon-theme = \"old\"\ntheme = \"new\"\n"},
		{name: "missing key", in: "[editor]\nline-number = \"relative\"\n", key: "theme", value: "new", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setTOMLString(tt.in, tt.key, tt.value)
			testutil.Equal(t, err != nil, tt.wantErr)
			if !tt.wantErr {
				testutil.Equal(t, got, tt.want)
			}
		})
	}
}

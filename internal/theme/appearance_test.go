package theme

import (
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/testutil"
)

func TestParseAppearance(t *testing.T) {
	tests := []struct {
		in      string
		want    Appearance
		wantErr bool
	}{
		{"dark", Dark, false},
		{"DARK", Dark, false},
		{"   DarK  ", Dark, false},
		{"", "", true},
		{"blue", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseAppearance(tt.in)
			testutil.Equal(t, got, tt.want)
			testutil.Equal(t, err != nil, tt.wantErr)
		})
	}
}

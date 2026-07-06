package integration

import (
	"github.com/nico-mayer/themectl-cli/internal/theme"
)

type SystemAppearance struct{}

func (SystemAppearance) Name() string {
	return "system-appearance"
}

func (SystemAppearance) Apply(t theme.Resolved) error {
	return setSystemAppearance(t.Appearance)
}

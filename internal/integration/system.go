package integration

import (
	"github.com/Nico-Mayer/themectl/internal/config"
	"github.com/Nico-Mayer/themectl/internal/theme"
)

type SystemAppearance struct{}

func (SystemAppearance) Name() string {
	return "system-appearance"
}

func (SystemAppearance) Check() error {
	return checkSystemAppearance()
}

func (SystemAppearance) Apply(t theme.Resolved) error {
	return setSystemAppearance(t.Appearance)
}

func newSystemAppearance(_ config.Config) Integration {
	return SystemAppearance{}
}

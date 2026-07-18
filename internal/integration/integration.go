package integration

import "github.com/Nico-Mayer/themectl-cli/internal/theme"

type Integration interface {
	Name() string
	Apply(t theme.Resolved) error
}

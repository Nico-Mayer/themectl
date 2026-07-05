package engine

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/nico-mayer/themectl-cli/internal/integration"
	"github.com/nico-mayer/themectl-cli/internal/theme"
)

type Engine struct {
	integrations []integration.Integration
}

func New(ints []integration.Integration) *Engine {
	return &Engine{
		integrations: ints,
	}
}

func (e *Engine) Apply(t theme.Resolved) error {
	var errs []error
	for _, in := range e.integrations {
		slog.Debug("applying integration", "integration", in.Name())
		if err := in.Apply(t); err != nil {
			slog.Warn("integration failed", "integration", in.Name(), "err", err)
			errs = append(errs, fmt.Errorf("%s: %w", in.Name(), err))
		}
	}
	return errors.Join(errs...)
}

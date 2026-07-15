package engine

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

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
	errs := make([]error, len(e.integrations))
	var wg sync.WaitGroup
	for i, in := range e.integrations {
		wg.Go(func() {
			slog.Debug("applying integration", "integration", in.Name())
			if err := in.Apply(t); err != nil {
				slog.Warn("integration failed", "integration", in.Name(), "err", err)
				errs[i] = fmt.Errorf("%s: %w", in.Name(), err)
			}
		})
	}
	wg.Wait()
	return errors.Join(errs...)
}

package integration

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

func ApplyAll(integrations []Integration, t theme.Resolved) error {
	warnErrors := make([]error, len(integrations))
	errs := make([]error, len(integrations))
	var wg sync.WaitGroup
	for i, in := range integrations {
		wg.Go(func() {
			if err := in.Check(); err != nil {
				warnErrors[i] = err
				return
			}
			slog.Debug("applying integration", "integration", in.Name())
			if err := in.Apply(t); err != nil {
				slog.Warn("integration failed", "integration", in.Name(), "err", err)
				errs[i] = fmt.Errorf("%s: %w", in.Name(), err)
			}
		})
	}
	wg.Wait()

	for i, err := range warnErrors {
		if err != nil {
			slog.Warn("integration unhealthy, skipping", "integration", integrations[i].Name(), "err", err)
		}
	}

	return errors.Join(errs...)
}

package engine

import (
	"errors"
	"testing"

	"github.com/Nico-Mayer/themectl-cli/internal/integration"
	"github.com/Nico-Mayer/themectl-cli/internal/theme"
)

type fakeIntegration struct {
	name    string
	err     error
	applied *bool
}

func (f fakeIntegration) Name() string {
	return f.name
}

func (f fakeIntegration) Apply(t theme.Resolved) error {
	if f.applied != nil {
		*f.applied = true
	}
	return f.err
}

func TestEngine_Apply_runsAllAndAggregatesErrors(t *testing.T) {
	var ranA, ranC bool
	e := New([]integration.Integration{
		fakeIntegration{name: "a", applied: &ranA},
		fakeIntegration{name: "b", err: errors.New("boom")},
		fakeIntegration{name: "c", applied: &ranC},
	})

	err := e.Apply(theme.Resolved{Family: "f", Variant: "v"})
	if err == nil {
		t.Fatal("want aggregated error from failing integration, got nil")
	}
	if !ranA || !ranC {
		t.Errorf("a failing integration must not stop the others (a=%v c=%v)", ranA, ranC)
	}
}

package engine

import (
	"errors"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/integration"
	"github.com/Nico-Mayer/themectl/internal/theme"
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

type fakeCheckedIntegration struct {
	fakeIntegration
	checkErr error
}

func (f fakeCheckedIntegration) Check() error {
	return f.checkErr
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

func TestEngine_Apply_skipsUnhealthyIntegrations(t *testing.T) {
	var ranBroken, ranHealthy, ranUnchecked bool
	e := New([]integration.Integration{
		fakeCheckedIntegration{fakeIntegration{name: "broken", applied: &ranBroken}, errors.New("config dir missing")},
		fakeCheckedIntegration{fakeIntegration{name: "healthy", applied: &ranHealthy}, nil},
		fakeIntegration{name: "unchecked", applied: &ranUnchecked},
	})

	err := e.Apply(theme.Resolved{Family: "f", Variant: "v"})
	if err == nil {
		t.Fatal("want error from unhealthy integration, got nil")
	}
	if ranBroken {
		t.Error("unhealthy integration must not be applied")
	}
	if !ranHealthy || !ranUnchecked {
		t.Errorf("healthy integrations must still run (healthy=%v unchecked=%v)", ranHealthy, ranUnchecked)
	}
}

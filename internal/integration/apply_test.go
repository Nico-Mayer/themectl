package integration

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

type fakeIntegration struct {
	name        string
	err         error
	checkErr    error
	unsupported bool
	applied     *bool
}

func (f fakeIntegration) Name() string {
	return f.name
}

func (f fakeIntegration) Check() error {
	return f.checkErr
}

func (f fakeIntegration) Supports(theme.Resolved) bool {
	return !f.unsupported
}

func (f fakeIntegration) Apply(t theme.Resolved) error {
	if f.applied != nil {
		*f.applied = true
	}
	return f.err
}

func TestEngine_ApplyAll_runsAllAndAggregatesErrors(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	var ranA, ranC bool
	integrations := []Integration{
		fakeIntegration{name: "a", applied: &ranA},
		fakeIntegration{name: "b", err: errors.New("boom")},
		fakeIntegration{name: "c", applied: &ranC},
	}

	err := ApplyAll(integrations, theme.Resolved{Family: "f", Variant: "v"})
	if err == nil {
		t.Fatal("want aggregated error from failing integration, got nil")
	}
	if !ranA || !ranC {
		t.Errorf("a failing integration must not stop the others (a=%v c=%v)", ranA, ranC)
	}
	if strings.Contains(logs.String(), "integration failed") {
		t.Errorf("returned integration error was also logged: %s", logs.String())
	}
}

func TestEngine_Apply_skipsUnsupportedIntegrations(t *testing.T) {
	var ranUnsupported, ranSupported bool
	integrations := []Integration{
		fakeIntegration{name: "unsupported", applied: &ranUnsupported, unsupported: true},
		fakeIntegration{name: "supported", applied: &ranSupported},
	}

	err := ApplyAll(integrations, theme.Resolved{Family: "f", Variant: "v"})
	if err != nil {
		t.Fatalf("unsupported integration must be skipped silently, not fail apply: %v", err)
	}
	if ranUnsupported {
		t.Error("unsupported integration must not be applied")
	}
	if !ranSupported {
		t.Error("supported integration must still run")
	}
}

func TestEngine_Apply_skipsUnhealthyIntegrations(t *testing.T) {
	var ranBroken, ranHealthy bool
	integrations := []Integration{
		fakeIntegration{name: "broken", applied: &ranBroken, checkErr: errors.New("config dir missing")},
		fakeIntegration{name: "healthy", applied: &ranHealthy},
	}

	err := ApplyAll(integrations, theme.Resolved{Family: "f", Variant: "v"})
	if err != nil {
		t.Fatalf("unhealthy integration must only warn, not fail apply: %v", err)
	}
	if ranBroken {
		t.Error("unhealthy integration must not be applied")
	}
	if !ranHealthy {
		t.Error("healthy integration must still run")
	}
}

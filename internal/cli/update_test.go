package cli

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/Nico-Mayer/themectl/internal/store"
)

func TestUpdateConfirmation_DisclosesConsequence(t *testing.T) {
	got := updateConfirmation("custom")
	for _, text := range []string{"Theme family \"custom\"", "local changes", "may discard them", "Continue?"} {
		if !strings.Contains(got, text) {
			t.Errorf("confirmation %q missing %q", got, text)
		}
	}
}

func TestLogUpdateResult_CommunicatesOutcome(t *testing.T) {
	tests := []struct {
		name   string
		result store.UpdateResult
		want   []string
	}{
		{name: "updated", result: store.UpdateResult{Name: "one", Status: store.UpdateUpdated}, want: []string{"theme family updated", "family=one"}},
		{name: "declined", result: store.UpdateResult{Name: "two", Status: store.UpdateDeclined}, want: []string{"theme family skipped", "family=two", "reason=\"local changes kept\""}},
		{name: "confirmation failed", result: store.UpdateResult{Name: "two", Status: store.UpdateDeclined, Err: errors.New("terminal unavailable")}, want: []string{"theme family skipped", "family=two", "reason=\"confirmation failed\"", "terminal unavailable"}},
		{name: "not git", result: store.UpdateResult{Name: "three", Status: store.UpdateSkipped}, want: []string{"theme family skipped", "family=three", "reason=\"not installed from Git\""}},
		{name: "failed", result: store.UpdateResult{Name: "four", Status: store.UpdateFailed, Err: errors.New("network down")}, want: []string{"theme family update failed", "family=four", "network down"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug})))
			t.Cleanup(func() { slog.SetDefault(previous) })

			logUpdateResult(tt.result)
			for _, text := range tt.want {
				if !strings.Contains(out.String(), text) {
					t.Errorf("log %q missing %q", out.String(), text)
				}
			}
		})
	}
}

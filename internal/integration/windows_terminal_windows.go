//go:build windows

package integration

// The integration itself is portable file and JSON work, so only its
// registration is gated. Everything else stays testable on any host.
func init() { available["windows-terminal"] = newWindowsTerminal }

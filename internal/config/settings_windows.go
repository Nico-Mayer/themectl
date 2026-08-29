//go:build windows

package config

// Windows Terminal is the default terminal on Windows 11, so it is on by
// default there and absent everywhere else.
func init() { defaultIntegrations = append(defaultIntegrations, "windows-terminal") }

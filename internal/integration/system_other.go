//go:build !darwin && !windows

package integration

import (
	"fmt"
	"runtime"

	"github.com/Nico-Mayer/themectl/internal/theme"
)

func checkSystemAppearance() error {
	return fmt.Errorf("system appearance: unsupported os: %s", runtime.GOOS)
}

func setSystemAppearance(theme.Appearance) error {
	return checkSystemAppearance()
}

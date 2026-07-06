//go:build !darwin && !windows

package integration

import (
	"fmt"
	"runtime"

	"github.com/nico-mayer/themectl-cli/internal/theme"
)

func setSystemAppearance(theme.Appearance) error {
	return fmt.Errorf("system appearance: unsupported os: %s", runtime.GOOS)
}

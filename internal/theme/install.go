package theme

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

var familyNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

func Install(themesDir, url, name string, force bool) (string, error) {
	if !familyNamePattern.MatchString(name) {
		return "", errors.New("name not allowed")
	}

	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("themectl install requires the git CLI: %w", err)
	}

	return "", nil
}

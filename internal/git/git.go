package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Installed() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("themectl requires the git CLI: %w", err)
	}
	return nil
}

func Run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func RemoteHead(url string) (string, error) {
	out, err := Run("", "ls-remote", url, "HEAD")
	if err != nil {
		return "", fmt.Errorf("ls-remote %s: %w", url, err)
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "", fmt.Errorf("ls-remote %s: empty HEAD", url)
	}
	return fields[0], nil
}

func CloneShallow(url, dst string) error {
	if out, err := Run("", "clone", "--depth", "1", url, dst); err != nil {
		return fmt.Errorf("git clone %s: %w (%s)", url, err, out)
	}
	return nil
}

func SparseClone(url, dst string, dirs ...string) error {
	steps := [][]string{
		{"clone", "--depth", "1", "--filter=blob:none", "--no-checkout", url, dst},
		append([]string{"-C", dst, "sparse-checkout", "set"}, dirs...),
		{"-C", dst, "checkout"},
	}
	for _, args := range steps {
		if out, err := Run("", args...); err != nil {
			return fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, out)
		}
	}
	return nil
}

func NormalizeURL(url string) string {
	if strings.Contains(url, "://") {
		return url
	}
	return "https://" + url
}

package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Nico-Mayer/themectl/internal/git"
)

type UpdateStatus int

const (
	UpdateSkipped  UpdateStatus = iota // not a git repository
	UpdateUpdated                      // git pull succeeded
	UpdateDeclined                     // dirty, user said no
	UpdateFailed                       // git pull failed
)

type UpdateResult struct {
	Name   string
	Status UpdateStatus
	Err    error
}

func Update(themesDir string, confirm func(string) bool) ([]UpdateResult, error) {
	if err := git.Installed(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read themes directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			names = append(names, e.Name())
		}
	}

	type classified struct {
		isRepo bool
		dirty  bool
	}

	infos := make([]classified, len(names))

	var wg sync.WaitGroup

	for i, name := range names {
		wg.Go(func() {
			dir := filepath.Join(themesDir, name)
			if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
				return
			}
			out, err := git.Run(dir, "status", "--porcelain")
			infos[i] = classified{isRepo: true, dirty: err == nil && out != ""}
		})
	}
	wg.Wait()

	results := make([]UpdateResult, len(names))
	var toPull []int
	for i, name := range names {
		switch {
		case !infos[i].isRepo:
			results[i] = UpdateResult{Name: name, Status: UpdateSkipped}
		case infos[i].dirty && !confirm(name):
			results[i] = UpdateResult{Name: name, Status: UpdateDeclined}
		default:
			toPull = append(toPull, i)
		}
	}

	for _, i := range toPull {
		wg.Go(func() {
			name := names[i]
			if out, err := git.Run(filepath.Join(themesDir, name), "pull"); err != nil {
				results[i] = UpdateResult{Name: name, Status: UpdateFailed, Err: fmt.Errorf("git pull: %w (%s)", err, out)}
				return
			}
			results[i] = UpdateResult{Name: name, Status: UpdateUpdated}
		})
	}
	wg.Wait()
	return results, nil
}

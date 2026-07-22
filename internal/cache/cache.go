package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

type Cache struct {
	dir string
}

func New(dir string) Cache {
	return Cache{dir: dir}
}

func (c Cache) Put(key, data string) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.path(key), []byte(data), 0o644)
}

func (c Cache) Get(key string) (string, bool) {
	data, err := os.ReadFile(c.path(key))
	if err != nil {
		return "", false
	}

	return string(data), true
}

func (c Cache) Fresh(key string, ttl time.Duration) bool {
	info, err := os.Stat(c.path(key))
	if err != nil {
		return false
	}

	if time.Since(info.ModTime()) > ttl {
		return false
	}

	return true
}

func (c Cache) Touch(key string) error {
	now := time.Now()
	return os.Chtimes(c.path(key), now, now)
}

func (c Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

func (c Cache) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, hex.EncodeToString(sum[:8]))
}

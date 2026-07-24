package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Nico-Mayer/themectl/internal/cache"
)

const DefaultMaxSize = 1024 * 1024 * 10

type Fetcher struct {
	Client  *http.Client
	Cache   cache.Cache
	TTL     time.Duration
	MaxSize uint64
}

func NewFetcher(client *http.Client, cache cache.Cache, ttl time.Duration) *Fetcher {
	return &Fetcher{
		Client:  client,
		Cache:   cache,
		TTL:     ttl,
		MaxSize: DefaultMaxSize,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	if !strings.HasPrefix(rawURL, "https://") {
		return nil, fmt.Errorf("invalid url %q: https required", rawURL)
	}

	if f.Cache.Fresh(rawURL, f.TTL) {
		data, ok := f.Cache.Get(rawURL)
		if ok {
			return data, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return f.staleFallback(rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return f.staleFallback(rawURL, fmt.Errorf("fetch %q: %s", rawURL, resp.Status))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) > int(f.MaxSize) {
		return nil, fmt.Errorf("response exceeds max size %d bytes", f.MaxSize)
	}

	if err := f.Cache.Put(rawURL, data); err != nil {
		return nil, err
	}

	return data, nil
}

func (f *Fetcher) staleFallback(url string, err error) ([]byte, error) {
	stale, ok := f.Cache.Get(url)
	if !ok {
		return nil, err
	}
	return stale, nil
}

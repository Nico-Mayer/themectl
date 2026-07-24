package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Nico-Mayer/themectl/internal/cache"
	"github.com/Nico-Mayer/themectl/internal/testutil"
)

const testTTL = time.Hour

const deadURL = "https://127.0.0.1:1/nvim.lua"

type testServer struct {
	*httptest.Server
	hits atomic.Int64
}

func (ts *testServer) Hits() int {
	return int(ts.hits.Load())
}

func newTestServer(t *testing.T, routes map[string]string) *testServer {
	t.Helper()

	ts := &testServer{}
	ts.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.hits.Add(1)
		body, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)

	return ts
}

func newCache(t *testing.T) cache.Cache {
	t.Helper()
	return cache.New(filepath.Join(t.TempDir(), "web"))
}

func TestFetchColdPopulatesCache(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "remote"})
	url := ts.URL + "/nvim.lua"
	dir := filepath.Join(t.TempDir(), "web")

	f := NewFetcher(ts.Client(), cache.New(dir), testTTL)
	got, err := f.Fetch(context.Background(), url)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "remote")

	ts.Close()

	f2 := NewFetcher(ts.Client(), cache.New(dir), testTTL)
	got, err = f2.Fetch(context.Background(), url)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "remote")
}

func TestFetchFreshCacheSkipsNetwork(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "remote"})
	url := ts.URL + "/nvim.lua"

	c := newCache(t)
	testutil.NoErr(t, c.Put(url, []byte("cached")))

	f := NewFetcher(ts.Client(), c, testTTL)
	got, err := f.Fetch(context.Background(), url)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "cached")
	testutil.Equal(t, ts.Hits(), 0)
}

func TestFetchStaleDeadServerFallsBack(t *testing.T) {
	c := newCache(t)
	testutil.NoErr(t, c.Put(deadURL, []byte("stale")))

	f := NewFetcher(&http.Client{Timeout: time.Second}, c, 0)
	got, err := f.Fetch(context.Background(), deadURL)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "stale")
}

func TestFetchStaleLiveServerRefetches(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "new"})
	url := ts.URL + "/nvim.lua"

	c := newCache(t)
	testutil.NoErr(t, c.Put(url, []byte("old")))

	f := NewFetcher(ts.Client(), c, 0)
	got, err := f.Fetch(context.Background(), url)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "new")

	cached, ok := c.Get(url)
	testutil.Equal(t, ok, true)
	testutil.Equal(t, string(cached), "new")
}

func TestFetchNothingCachedDeadServerErrors(t *testing.T) {
	f := NewFetcher(&http.Client{Timeout: time.Second}, newCache(t), testTTL)
	got, err := f.Fetch(context.Background(), deadURL)
	if err == nil {
		t.Fatalf("want error for dead server with cold cache, got %q", got)
	}
}

func TestFetchNon200(t *testing.T) {
	t.Run("nothing cached errors", func(t *testing.T) {
		ts := newTestServer(t, nil)
		f := NewFetcher(ts.Client(), newCache(t), testTTL)

		got, err := f.Fetch(context.Background(), ts.URL+"/missing.lua")
		if err == nil {
			t.Fatalf("want error for 404 with cold cache, got %q", got)
		}
	})

	t.Run("stale cache falls back", func(t *testing.T) {
		ts := newTestServer(t, nil)
		url := ts.URL + "/missing.lua"

		c := newCache(t)
		testutil.NoErr(t, c.Put(url, []byte("stale")))

		f := NewFetcher(ts.Client(), c, 0)
		got, err := f.Fetch(context.Background(), url)
		testutil.NoErr(t, err)
		testutil.Equal(t, string(got), "stale")
	})
}

func TestFetchRejectsBadScheme(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "remote"})
	plain := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.hits.Add(1)
	}))
	t.Cleanup(plain.Close)

	f := NewFetcher(ts.Client(), newCache(t), testTTL)
	for _, url := range []string{plain.URL + "/nvim.lua", "file:///etc/passwd"} {
		if _, err := f.Fetch(context.Background(), url); err == nil {
			t.Errorf("want error for %s, got nil", url)
		}
	}
	testutil.Equal(t, ts.Hits(), 0)
}

func TestFetchHonorsContext(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "remote"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := NewFetcher(ts.Client(), newCache(t), testTTL)
	if got, err := f.Fetch(ctx, ts.URL+"/nvim.lua"); err == nil {
		t.Fatalf("want error for canceled context, got %q", got)
	}
}

func TestFetchFreshCacheUnreadableRefetches(t *testing.T) {
	ts := newTestServer(t, map[string]string{"/nvim.lua": "remote"})
	url := ts.URL + "/nvim.lua"

	dir := filepath.Join(t.TempDir(), "web")
	c := cache.New(dir)
	testutil.NoErr(t, c.Put(url, []byte("cached")))

	entries, err := filepath.Glob(filepath.Join(dir, "*"))
	testutil.NoErr(t, err)
	testutil.Equal(t, len(entries), 1)
	testutil.NoErr(t, os.Chmod(entries[0], 0o200))

	f := NewFetcher(ts.Client(), c, testTTL)
	got, err := f.Fetch(context.Background(), url)
	testutil.NoErr(t, err)
	testutil.Equal(t, string(got), "remote")
}

func TestFetchSizeCap(t *testing.T) {
	url := "/big.lua"
	ts := newTestServer(t, map[string]string{url: strings.Repeat("x", 1024)})

	c := newCache(t)
	f := NewFetcher(ts.Client(), c, testTTL)
	f.MaxSize = 64

	got, err := f.Fetch(context.Background(), ts.URL+url)
	if err == nil {
		t.Fatalf("want error for response over MaxSize, got %d bytes", len(got))
	}

	if data, ok := c.Get(ts.URL + url); ok && len(data) > 64 {
		t.Errorf("oversized data written to cache: %d bytes", len(data))
	}
}

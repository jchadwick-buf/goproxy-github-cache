package main

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/goproxy/goproxy"
	"github.com/pkg/errors"
	actionscache "github.com/tonistiigi/go-actions-cache"
)

const githubCacheKeyPrefix = "goproxy-github-cache-"

// githubCacher is a goproxy.Cacher implementation for GitHub Actions cache.
type githubCacher struct {
	client *http.Client
	github *actionscache.Cache
}

// newGithubCacher constructs a new cacher using GitHub Actions cache
func newGithubCacher(client *http.Client, github *actionscache.Cache) goproxy.Cacher {
	return &githubCacher{
		client: client,
		github: github,
	}
}

// Get implements goproxy.Cacher.
func (cache *githubCacher) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	log.Printf("loading key: %v\n", name)
	entry, err := cache.github.Load(ctx, githubCacheKey(name))
	if err != nil {
		log.Printf("error loading cache: %q", name)
		return nil, err
	}
	if entry == nil {
		return nil, fs.ErrNotExist
	}
	return githubCacheDownload(ctx, cache.client, entry)
}

// Put implements goproxy.Cacher.
func (cache *githubCacher) Put(ctx context.Context, name string, content io.ReadSeeker) error {
	log.Printf("saving key: %v\n", name)
	buffer := &bytes.Buffer{}
	if _, err := io.Copy(buffer, content); err != nil {
		log.Printf("error saving cache: %q: %v", name, err)
		return nil
	}
	if err := cache.github.Save(ctx, githubCacheKey(name), &nopCloserByteReader{*bytes.NewReader(buffer.Bytes())}); err != nil {
		log.Printf("error saving cache %q: %v", name, err)
	}
	return nil
}

type nopCloserByteReader struct {
	bytes.Reader
}

func (nopCloserByteReader) Close() error {
	return nil
}

func githubCacheKey(name string) string {
	return githubCacheKeyPrefix + name
}

func githubCacheDownload(ctx context.Context, client *http.Client, ce *actionscache.Entry) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", ce.URL, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req = req.WithContext(ctx)
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
			return nil, errors.Errorf("invalid status response %v for %s, range: %v", resp.Status, ce.URL, req.Header.Get("Range"))
		}
		return nil, errors.Errorf("invalid status response %v for %s", resp.Status, ce.URL)
	}
	return resp.Body, nil
}

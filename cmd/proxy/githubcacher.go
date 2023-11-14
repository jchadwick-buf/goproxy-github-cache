package main

import (
	"context"
	"io"
	"io/fs"
	"log"

	"github.com/goproxy/goproxy"
	actionscache "github.com/tonistiigi/go-actions-cache"
)

const githubCacheKeyPrefix = "goproxy-github-cache-"

// githubCacher is a goproxy.Cacher implementation for GitHub Actions cache.
type githubCacher struct {
	github *actionscache.Cache
}

// newGithubCacher constructs a new cacher using GitHub Actions cache
func newGithubCacher(github *actionscache.Cache) goproxy.Cacher {
	return &githubCacher{github: github}
}

// Get implements goproxy.Cacher.
func (cache *githubCacher) Get(ctx context.Context, name string) (io.ReadCloser, error) {
	entry, err := cache.github.Load(ctx, githubCacheKey(name))
	if err != nil {
		log.Printf("error loading cache: %q", name)
		return nil, err
	}
	if entry == nil {
		return nil, fs.ErrNotExist
	}
	return newReadCursor(entry.Download(ctx)), nil
}

// Put implements goproxy.Cacher.
func (cache *githubCacher) Put(ctx context.Context, name string, content io.ReadSeeker) error {
	blob, err := newReaderBlob(content)
	if err != nil {
		log.Printf("error saving cache: %q", name)
		return err
	}
	return cache.github.Save(ctx, githubCacheKey(name), blob)
}

func githubCacheKey(name string) string {
	return githubCacheKeyPrefix + name
}

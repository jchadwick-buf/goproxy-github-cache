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
	log.Printf("loading key: %v\n", name)
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
	log.Printf("saving key: %v\n", name)
	blob, err := newReaderBlob(content)
	if err != nil {
		log.Printf("error saving cache: %q: %v", name, err)
		return nil
	}
	if err := cache.github.Save(ctx, githubCacheKey(name), blob); err != nil {
		log.Printf("error saving cache %q: %v", name, err)
	}
	return nil
}

func githubCacheKey(name string) string {
	return githubCacheKeyPrefix + name
}

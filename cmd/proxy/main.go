package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/bufbuild/httplb"
	"github.com/goproxy/goproxy"
	actionscache "github.com/tonistiigi/go-actions-cache"
)

func main() {
	listen := flag.String("listen", ":8123", "address to listen on")
	flag.Parse()
	client := httplb.NewClient()
	cache, err := actionscache.TryEnv(actionscache.Opt{
		Client:      client.Client,
		Timeout:     10 * time.Second,
		BackoffPool: &actionscache.BackoffPool{},
	})
	if err != nil {
		log.Fatalf("error initializing actions cache: %v", err)
	}
	proxy := &goproxy.Goproxy{
		Cacher:    newGithubCacher(cache),
		Transport: client.Transport,
	}
	server := &http.Server{
		Addr:    *listen,
		Handler: proxy,
	}
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("error serving goproxy: %v", err)
	}
}

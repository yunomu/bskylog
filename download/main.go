package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/yunomu/bskylog/consumer"
	"github.com/yunomu/bskylog/processor"
	"github.com/yunomu/bskylog/scanner"
)

var (
	handle   = flag.String("handle", "", "your handle xxx.bsky.social")
	password = flag.String("password", "", "xxxx-xxxx-xxxx-xxxx")
	host     = flag.String("host", "https://bsky.social", "bsky host")
)

func init() {
	flag.Parse()
}

func main() {
	if *handle == "" {
		slog.Error("Handle is empty")
		return
	}

	if *password == "" {
		slog.Error("password is empty")
		return
	}

	ctx := context.Background()

	client := &xrpc.Client{
		Host: *host,
	}

	auth, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: *handle,
		Password:   *password,
	})
	if err != nil {
		slog.Error("Authentication error",
			"err", err,
			"handle", *handle,
			"password", *password,
		)
		return
	}
	slog.Info("Auth",
		"handle", auth.Handle,
		"auth", auth,
	)

	client.Auth = &xrpc.AuthInfo{
		AccessJwt:  auth.AccessJwt,
		RefreshJwt: auth.RefreshJwt,
		Did:        auth.Did,
		Handle:     auth.Handle,
	}

	filename := "posts.json"
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("Create file",
			"err", err,
			"filename", filename,
		)
		return
	}
	defer file.Close()

	p := processor.New(
		scanner.NewXRPCScanner(
			client,
			auth.Did,
			scanner.SetLogger(slog.With("module", "scanner")),
		),
		consumer.NewJSONRecord(
			file,
		),
	)
	defer p.Close()

	slog.Debug("Started retrieving posts...")
	if err := p.Proc(ctx); err != nil {
		slog.Error("Proc", "err", err)
	}

	slog.Info("Complete", "file", filename)
}

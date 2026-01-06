package main

import (
	"context"
	"flag"
	"log/slog"
	"time"

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

	dir  = flag.String("dir", "output", "output directory")
	zone = flag.String("zone", "Asia/Tokyo", "Time zone")
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

	loc, err := time.LoadLocation(*zone)
	if err != nil {
		slog.Error("LoadLocation", "zone", *zone)
		return
	}

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

	p := processor.New(
		scanner.NewXRPCScanner(
			client,
			auth.Did,
			scanner.SetLogger(slog.With("module", "scanner")),
		),
		consumer.NewDailyJSONRecord(
			*dir,
			consumer.SetDailyJSONRecordLocation(loc),
			consumer.SetDailyJSONRecordLogger(slog.With("module", "consumer")),
		),
	)
	defer p.Close()

	slog.Debug("Started retrieving posts...")
	if err := p.Proc(ctx); err != nil {
		slog.Error("Proc", "err", err)
	}

	slog.Info("Complete", "output_directory", *dir)
}

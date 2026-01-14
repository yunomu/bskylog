package bsky

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/yunomu/bskylog/cmd/bsky/archive"
	"github.com/yunomu/bskylog/cmd/bsky/download"
)

type command struct {
	handle   *string
	password *string
	host     *string

	commander *subcommands.Commander
}

func (c *command) Name() string     { return "bsky" }
func (c *command) Synopsis() string { return "bsky command" }
func (c *command) Usage() string {
	return `bsky [options] <subcommand>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.handle = f.String("handle", "", "your handle xxx.bsky.social")
	c.password = f.String("password", "", "xxxx-xxxx-xxxx-xxxx")
	c.host = f.String("host", "https://bsky.social", "bsky host")

	commander := subcommands.NewCommander(f, "bsky")
	commander.Register(download.NewCommand(), "")
	commander.Register(archive.NewCommand(), "")
	c.commander = commander
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.handle == "" {
		slog.Error("Handle is empty")
		return subcommands.ExitFailure
	}

	if *c.password == "" {
		slog.Error("password is empty")
		return subcommands.ExitFailure
	}

	client := &xrpc.Client{
		Host: *c.host,
	}

	auth, err := atproto.ServerCreateSession(ctx, client, &atproto.ServerCreateSession_Input{
		Identifier: *c.handle,
		Password:   *c.password,
	})
	if err != nil {
		slog.Error("Authentication error",
			"err", err,
			"handle", *c.handle,
			"password", *c.password,
		)
		return subcommands.ExitFailure
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

	return c.commander.Execute(ctx, client, auth)
}

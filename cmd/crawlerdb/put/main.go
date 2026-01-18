package put

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/google/subcommands"
	"github.com/yunomu/bskylog/lib/crawlerdb"
)

type command struct {
	did       *string
	latest    *string
	timestamp *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "put" }
func (c *command) Synopsis() string { return "put" }
func (c *command) Usage() string {
	return `put -did {did} -latest {latest} -timestamp {timestamp}
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.did = f.String("did", "", "Did")
	c.latest = f.String("latest", "", "Latest")
	c.timestamp = f.String("timestamp", "", "Timestamp (RFC3339)")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.did == "" {
		slog.Error("-did is required")
		return subcommands.ExitFailure
	}
	if *c.latest == "" {
		slog.Error("-latest is required")
		return subcommands.ExitFailure
	}
	if *c.timestamp == "" {
		slog.Error("-timestamp is required")
		return subcommands.ExitFailure
	}

	ts, err := time.Parse(time.RFC3339Nano, *c.timestamp)
	if err != nil {
		slog.Error("time.Parse", "err", err, "timestamp", *c.timestamp)
		return subcommands.ExitFailure
	}

	if len(args) == 0 {
		slog.Error("db not found")
		return subcommands.ExitFailure
	}
	client, ok := args[0].(crawlerdb.DB)
	if !ok {
		slog.Error("unexpected type", "arg", args[0])
		return subcommands.ExitFailure
	}

	if err := client.Put(ctx, &crawlerdb.Timestamp{
		Did:       *c.did,
		LatestCid: *c.latest,
		Timestamp: ts.Unix(),
	}); err != nil {
		slog.Error("Put", "err", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

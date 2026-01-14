package archive

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/google/subcommands"

	"github.com/yunomu/bskylog/lib/consumer"
	"github.com/yunomu/bskylog/lib/processor"
	"github.com/yunomu/bskylog/lib/scanner"
)

type command struct {
	dir  *string
	zone *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "archive" }
func (c *command) Synopsis() string { return "archive [options]" }
func (c *command) Usage() string {
	return `archive [options]
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.dir = f.String("dir", "output", "output directory")
	c.zone = f.String("zone", "Asia/Tokyo", "Time zone")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	client := args[0].(*xrpc.Client)
	auth := args[1].(*atproto.ServerCreateSession_Output)

	loc, err := time.LoadLocation(*c.zone)
	if err != nil {
		slog.Error("LoadLocation", "zone", *c.zone)
		return subcommands.ExitFailure
	}

	p := processor.New(
		scanner.NewXRPCScanner(
			client,
			auth.Did,
			scanner.SetLogger(slog.With("module", "scanner")),
		),
		consumer.NewDailyJSONRecord(
			*c.dir,
			consumer.SetDailyJSONRecordLocation(loc),
			consumer.SetDailyJSONRecordLogger(slog.With("module", "consumer")),
		),
	)
	defer p.Close()

	slog.Debug("Started retrieving posts...")
	if err := p.Proc(ctx); err != nil {
		slog.Error("Proc", "err", err)
		return subcommands.ExitFailure
	}

	slog.Info("Complete", "output_directory", *c.dir)
	return subcommands.ExitSuccess
}

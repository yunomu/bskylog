package download

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/google/subcommands"

	"github.com/yunomu/bskylog/lib/consumer"
	"github.com/yunomu/bskylog/lib/processor"
	"github.com/yunomu/bskylog/lib/scanner"
)

type command struct {
	filename *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "download" }
func (c *command) Synopsis() string { return "download" }
func (c *command) Usage() string {
	return `download
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.filename = f.String("filename", "post.json", "Output filename")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) < 2 {
		slog.Error("arguments not found")
		return subcommands.ExitFailure
	}
	client, ok := args[0].(*xrpc.Client)
	if !ok {
		slog.Error("unexpected type", "arg", args[0])
		return subcommands.ExitFailure
	}
	auth, ok := args[1].(*atproto.ServerCreateSession_Output)
	if !ok {
		slog.Error("unexpected type", "arg", args[1])
		return subcommands.ExitFailure
	}

	file, err := os.Create(*c.filename)
	if err != nil {
		slog.Error("Create file",
			"err", err,
			"filename", *c.filename,
		)
		return subcommands.ExitFailure
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
	defer p.Close(ctx)

	slog.Debug("Started retrieving posts...")
	if err := p.Proc(ctx); err != nil {
		slog.Error("Proc", "err", err)
		return subcommands.ExitFailure
	}

	slog.Info("Complete", "file", *c.filename)
	return subcommands.ExitSuccess
}

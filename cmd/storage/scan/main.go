package scan

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/subcommands"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/yunomu/bskylog/lib/storage"
)

type command struct{}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "scan" }
func (c *command) Synopsis() string { return "scan posts from storage" }
func (c *command) Usage() string {
	return `scan:
  Scan posts from storage and output in TSV format.
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	// No flags for this subcommand
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) < 1 {
		slog.Error("scanner not found in args")
		return subcommands.ExitFailure
	}
	scanner, ok := args[0].(storage.Scanner)
	if !ok {
		slog.Error("unexpected type for scanner", "arg", args[0])
		return subcommands.ExitFailure
	}

	if err := scanner.Scan(ctx, func(key string, position int, post *bsky.FeedDefs_FeedViewPost) error {
		postJSON, err := json.Marshal(post)
		if err != nil {
			slog.Error("failed to marshal post to JSON", "err", err)
			return err
		}
		_, err = fmt.Fprintf(os.Stdout, "%s\t%d\t%s\n", key, position, postJSON)
		if err != nil {
			slog.Error("failed to write to stdout", "err", err)
			return err
		}
		return nil
	}); err != nil {
		slog.Error("failed to scan", "err", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

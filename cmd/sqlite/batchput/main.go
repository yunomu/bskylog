package batchput

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/google/subcommands"
	"gorm.io/gorm"

	"github.com/bluesky-social/indigo/api/bsky"

	"github.com/yunomu/bskylog/lib/index"
)

type command struct{}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "batchput" }
func (c *command) Synopsis() string { return "batch put posts to sqlite" }
func (c *command) Usage() string {
	return `batchput:
  Read TSV from stdin and batch put posts to sqlite.
  Format: key\tposition\tpost(JSON)
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	// No flags for this subcommand
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) < 1 {
		slog.Error("gorm.DB not found in args")
		return subcommands.ExitFailure
	}
	db, ok := args[0].(*gorm.DB)
	if !ok {
		slog.Error("unexpected type for gorm.DB", "arg", args[0])
		return subcommands.ExitFailure
	}

	// Use transaction for batch insert
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		slog.Error("failed to begin transaction", "err", tx.Error)
		return subcommands.ExitFailure
	}

	recordStore := index.NewGorm(tx) // Pass the transaction to the record store

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			slog.Error("invalid TSV format", "line", line)
			tx.Rollback()
			return subcommands.ExitFailure
		}

		key := parts[0]
		position, err := strconv.Atoi(parts[1])
		if err != nil {
			slog.Error("invalid position format", "position", parts[1], "err", err)
			tx.Rollback()
			return subcommands.ExitFailure
		}

		var post bsky.FeedDefs_FeedViewPost
		if err := json.Unmarshal([]byte(parts[2]), &post); err != nil {
			slog.Error("failed to unmarshal post JSON", "json", parts[2], "err", err)
			tx.Rollback()
			return subcommands.ExitFailure
		}

		if err := recordStore.Put(ctx, key, position, &post); err != nil {
			slog.Error("failed to put record", "key", key, "position", position, "err", err)
			tx.Rollback()
			return subcommands.ExitFailure
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		slog.Error("error reading stdin", "err", err)
		tx.Rollback()
		return subcommands.ExitFailure
	}

	if err := tx.Commit().Error; err != nil {
		slog.Error("failed to commit transaction", "err", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

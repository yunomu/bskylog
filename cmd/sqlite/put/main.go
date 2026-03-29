package put

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"

	"github.com/google/subcommands"
	"gorm.io/gorm"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/yunomu/bskylog/lib/sqlite"
)

type command struct {
	key      *string
	position *int
	file     *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "put" }
func (c *command) Synopsis() string { return "put record to sqlite" }
func (c *command) Usage() string {
	return `put -key {key} -position {position}
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.key = f.String("key", "", "key")
	c.position = f.Int("pos", -1, "record position")
	c.file = f.String("file", "-", "input file path ('-' for stdin)")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) == 0 {
		slog.Error("db not found in args")
		return subcommands.ExitFailure
	}
	gormDB, ok := args[0].(*gorm.DB)
	if !ok {
		slog.Error("db has unexpected type")
		return subcommands.ExitFailure
	}

	if *c.key == "" {
		slog.Error("key is required")
		return subcommands.ExitFailure
	}
	if *c.position >= 0 {
		slog.Error("pos is required")
		return subcommands.ExitFailure
	}

	var reader io.Reader
	if *c.file == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(*c.file)
		if err != nil {
			slog.Error("failed to open file", "file", *c.file, "err", err)
			return subcommands.ExitFailure
		}
		defer file.Close()
		reader = file
	}

	decoder := json.NewDecoder(reader)
	var fvp bsky.FeedDefs_FeedViewPost
	if err := decoder.Decode(&fvp); err != nil {
		slog.Error("failed to decode JSON", "err", err)
		return subcommands.ExitFailure
	}

	rec := sqlite.ToRecord(*c.key, *c.position, &fvp)
	if rec == nil {
		slog.Error("failed to convert to sqlite.Record")
		return subcommands.ExitFailure
	}

	if err := gormDB.WithContext(ctx).Create(rec).Error; err != nil {
		slog.Error("failed to create record", "err", err)
		return subcommands.ExitFailure
	}

	slog.Info("record put successfully", "cid", rec.Cid)

	return subcommands.ExitSuccess
}

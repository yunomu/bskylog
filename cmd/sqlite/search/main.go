package search

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/subcommands"
	"gorm.io/gorm"

	"github.com/yunomu/bskylog/lib/index"
)

type command struct {
	query *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "search" }
func (c *command) Synopsis() string { return "Search records by text" }
func (c *command) Usage() string {
	return `search -query <search_query>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.query = f.String("query", "", "Search query string")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.query == "" {
		slog.Error("search query is empty")
		return subcommands.ExitUsageError
	}

	if len(args) == 0 {
		slog.Error("db not found in args")
		return subcommands.ExitFailure
	}
	db, ok := args[0].(*gorm.DB)
	if !ok {
		slog.Error("db has unexpected type")
		return subcommands.ExitFailure
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	gormIndex := index.NewGorm(db, index.GormOptionLogger(logger))

	results, err := gormIndex.Search(ctx, &index.Query{
		Text: strings.Fields(*c.query),
	})
	if err != nil {
		slog.Error("failed to execute search", "query", *c.query, "err", err)
		return subcommands.ExitFailure
	}

	for _, result := range results {
		fmt.Printf("Key: %s, Position: %d\n", result.Key, result.Position)
	}

	return subcommands.ExitSuccess
}

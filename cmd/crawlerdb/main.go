package crawlerdb

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/yunomu/bskylog/cmd/crawlerdb/list"
	"github.com/yunomu/bskylog/cmd/crawlerdb/put"
	"github.com/yunomu/bskylog/lib/crawlerdb"
)

type command struct {
	table *string

	commander *subcommands.Commander
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "crawlerdb" }
func (c *command) Synopsis() string { return "crawlerdb command" }
func (c *command) Usage() string {
	return `crawlerdb -table {table_name} <subcommand>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.table = f.String("table", "", "table name (CrawlerTable)")

	commander := subcommands.NewCommander(f, "bsky")
	commander.Register(list.NewCommand(), "")
	commander.Register(put.NewCommand(), "")
	c.commander = commander
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) == 0 {
		slog.Error("config not found in args")
		return subcommands.ExitFailure
	}
	cfg, ok := args[0].(map[string]string)
	if !ok {
		slog.Error("config has unexpected type")
		return subcommands.ExitFailure
	}

	table := *c.table
	if v, ok := cfg["CrawlerTable"]; ok && table == "" {
		table = v
	}
	if table == "" {
		slog.Error("table is empty")
		return subcommands.ExitFailure
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("LoadConfig", "error", err)
		return subcommands.ExitFailure
	}

	db := crawlerdb.NewDynamoDB(
		dynamodb.NewFromConfig(awsCfg),
		table,
	)

	return c.commander.Execute(ctx, db, cfg)
}

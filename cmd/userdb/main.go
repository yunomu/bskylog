package userdb

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/yunomu/bskylog/cmd/userdb/list"
	"github.com/yunomu/bskylog/cmd/userdb/put"
	"github.com/yunomu/bskylog/lib/userdb"
)

type command struct {
	table *string
	index *string

	commander *subcommands.Commander
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "userdb" }
func (c *command) Synopsis() string { return "userdb command" }
func (c *command) Usage() string {
	return `userdb -table {table_name} <subcommand>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.table = f.String("table", "", "table name (UserTable)")
	c.index = f.String("index", "", "Handle index name (HandleIndex)")

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
	if v, ok := cfg["UserTable"]; ok && table == "" {
		table = v
	}
	if table == "" {
		slog.Error("table is empty")
		return subcommands.ExitFailure
	}

	index := *c.index
	if v, ok := cfg["HandleIndex"]; ok && index == "" {
		index = v
	}
	if index == "" {
		slog.Error("Handle index is empty")
		return subcommands.ExitFailure
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("LoadConfig", "error", err)
		return subcommands.ExitFailure
	}

	db := userdb.NewDynamoDB(
		dynamodb.NewFromConfig(awsCfg),
		table,
		index,
	)

	return c.commander.Execute(ctx, db, cfg)
}

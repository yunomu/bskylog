package sqlite

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/yunomu/bskylog/cmd/sqlite/batchput"
	"github.com/yunomu/bskylog/cmd/sqlite/put"
	"github.com/yunomu/bskylog/cmd/sqlite/search" // searchパッケージをインポート
)

type command struct {
	dbpath *string

	commander *subcommands.Commander
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "sqlite" }
func (c *command) Synopsis() string { return "sqlite command" }
func (c *command) Usage() string {
	return `sqlite -dbpath {db_file_path} <subcommand>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.dbpath = f.String("dbpath", "bskylog.db", "SQLite database file path")

	commander := subcommands.NewCommander(f, "bsky")

	commander.Register(put.NewCommand(), "")
	commander.Register(batchput.NewCommand(), "")
	commander.Register(search.NewCommand(), "") // searchサブコマンドを登録
	c.commander = commander
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	dbpath := *c.dbpath
	if dbpath == "" {
		slog.Error("dbpath is empty")
		return subcommands.ExitFailure
	}

	db, err := gorm.Open(sqlite.Open(dbpath + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(10000)"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect database", "err", err)
		return subcommands.ExitFailure
	}

	return c.commander.Execute(ctx, db)
}

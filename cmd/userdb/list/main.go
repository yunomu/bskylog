package list

import (
	"context"
	"encoding/csv"
	"flag"
	"log/slog"
	"os"
	"strconv"

	"github.com/google/subcommands"
	"github.com/yunomu/bskylog/lib/userdb"
)

type command struct{}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "list" }
func (c *command) Synopsis() string { return "list" }
func (c *command) Usage() string {
	return `list
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) == 0 {
		slog.Error("db not found")
		return subcommands.ExitFailure
	}
	client, ok := args[0].(userdb.DB)
	if !ok {
		slog.Error("unexpected type", "arg", args[0])
		return subcommands.ExitFailure
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	w.Write([]string{"did", "handle", "password", "timezone"})

	if err := client.Scan(ctx, func(user *userdb.User) error {
		w.Write([]string{
			user.Did,
			user.Handle,
			user.Password,
			strconv.Itoa(user.TimeZone),
		})
		return nil
	}); err != nil {
		slog.Error("Scan error", "err", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

package put

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"
	"github.com/yunomu/bskylog/lib/userdb"
)

type command struct {
	did      *string
	handle   *string
	password *string
	timezone *int
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "put" }
func (c *command) Synopsis() string { return "put" }
func (c *command) Usage() string {
	return `put -did {did} -handle {handle} -password {password} -timezone {timezone}
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.did = f.String("did", "", "Did")
	c.handle = f.String("handle", "", "Handle")
	c.password = f.String("password", "", "Password")
	c.timezone = f.Int("timezone", 0, "Timezone (min)")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.did == "" {
		slog.Error("-did is required")
		return subcommands.ExitFailure
	}
	if *c.handle == "" {
		slog.Error("-handle is required")
		return subcommands.ExitFailure
	}
	if *c.password == "" {
		slog.Error("-password is required")
		return subcommands.ExitFailure
	}

	if len(args) == 0 {
		slog.Error("db not found")
		return subcommands.ExitFailure
	}
	client, ok := args[0].(userdb.DB)
	if !ok {
		slog.Error("unexpected type", "arg", args[0])
		return subcommands.ExitFailure
	}

	if err := client.Put(ctx, &userdb.User{
		Did:      *c.did,
		Handle:   *c.handle,
		Password: *c.password,
		TimeZone: *c.timezone,
	}); err != nil {
		slog.Error("Put", "err", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

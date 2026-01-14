package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	//"github.com/yunomu/blog/cmd/user"
)

var (
	configPath = flag.String("config", ".config", "config file path")
)

func init() {
	//subcommands.Register(image.NewCommand(), "")

	subcommands.Register(subcommands.CommandsCommand(), "other")
	subcommands.Register(subcommands.FlagsCommand(), "other")
	subcommands.Register(subcommands.HelpCommand(), "other")

	flag.Parse()
}

func main() {
	ctx := context.Background()

	os.Exit(int(subcommands.Execute(ctx, cfg)))
}

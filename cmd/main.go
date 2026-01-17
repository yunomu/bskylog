package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/google/subcommands"

	"github.com/yunomu/bskylog/cmd/bsky"
	"github.com/yunomu/bskylog/cmd/config"
	"github.com/yunomu/bskylog/cmd/userdb"
)

var (
	configPath = flag.String("config", ".config", "config file path")
)

func init() {
	subcommands.Register(userdb.NewCommand(), "")
	subcommands.Register(bsky.NewCommand(), "")
	subcommands.Register(config.NewCommand(), "")

	subcommands.Register(subcommands.CommandsCommand(), "other")
	subcommands.Register(subcommands.FlagsCommand(), "other")
	subcommands.Register(subcommands.HelpCommand(), "other")

	flag.Parse()
}

func main() {
	ctx := context.Background()

	cfg := make(map[string]string)
	if f, err := os.Open(*configPath); err != nil {
		slog.Warn("config file open error, ignore", "path", *configPath, "err", err)
	} else {
		defer f.Close()
		dec := json.NewDecoder(f)
		if err := dec.Decode(&cfg); err != nil {
			slog.Error("config file decode error", "path", *configPath, "err", err)
			os.Exit(1)
		}
	}

	os.Exit(int(subcommands.Execute(ctx, cfg)))
}

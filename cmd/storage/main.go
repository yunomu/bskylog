package storage

import (
	"context"
	"flag"
	"log/slog"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/yunomu/bskylog/cmd/storage/scan"
	"github.com/yunomu/bskylog/lib/storage"
)

type command struct {
	prefix *string

	commander *subcommands.Commander
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "storage" }
func (c *command) Synopsis() string { return "storage command" }
func (c *command) Usage() string {
	return `storage [-prefix {prefix}] <subcommand>
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.prefix = f.String("prefix", "", "S3 object prefix")

	commander := subcommands.NewCommander(f, "storage")
	commander.Register(scan.NewCommand(), "")
	c.commander = commander
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	// Check if the subcommand being executed is "flags" or "help"
	if len(f.Args()) > 0 {
		subCmdName := f.Args()[0]
		if subCmdName == "flags" || subCmdName == "help" {
			return c.commander.Execute(ctx, nil) // Pass nil for scanner as it's not needed for help
		}
	}

	if len(args) < 1 {
		slog.Error("config not found in args")
		return subcommands.ExitFailure
	}
	cfg, ok := args[0].(map[string]string)
	if !ok {
		slog.Error("config has unexpected type")
		return subcommands.ExitFailure
	}

	bucket := cfg["PublishBucket"]
	if bucket == "" {
		slog.Error("PublishBucket is empty")
		return subcommands.ExitFailure
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("LoadConfig", "err", err)
		return subcommands.ExitFailure
	}

	s3Client := s3.NewFromConfig(awsCfg)
	
	var scanner storage.Scanner
	if *c.prefix != "" {
		scanner = storage.NewS3(s3Client, bucket, storage.S3OptionPrefix(*c.prefix))
	} else {
		scanner = storage.NewS3(s3Client, bucket)
	}

	return c.commander.Execute(ctx, scanner)
}

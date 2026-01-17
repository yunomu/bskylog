package config

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"

	"github.com/google/subcommands"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type command struct {
	stack *string
}

func NewCommand() subcommands.Command {
	return &command{}
}

func (c *command) Name() string     { return "config" }
func (c *command) Synopsis() string { return "stack config" }
func (c *command) Usage() string {
	return `config -stack {stack_name}
`
}

func (c *command) SetFlags(f *flag.FlagSet) {
	c.stack = f.String("stack", "", "stack name")
}

func (c *command) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if *c.stack == "" {
		slog.Error("stack is empty")
		return subcommands.ExitFailure
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		slog.Error("LoadConfig", "error", err)
		return subcommands.ExitFailure
	}

	client := cloudformation.NewFromConfig(cfg)

	ret, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(*c.stack),
	})
	if err != nil {
		slog.Error("DescribeStacks", "error", err, "stackName", *c.stack)
		return subcommands.ExitFailure
	}

	out := make(map[string]string)
	for _, stack := range ret.Stacks {
		for _, param := range stack.Parameters {
			out[aws.ToString(param.ParameterKey)] = aws.ToString(param.ParameterValue)
		}

		for _, kv := range stack.Outputs {
			out[aws.ToString(kv.OutputKey)] = aws.ToString(kv.OutputValue)
		}

		break
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(out); err != nil {
		slog.Error("json.Encode", "err", err, "value", out)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

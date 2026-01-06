package processor

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"

	"github.com/yunomu/bskylog/consumer"
	"github.com/yunomu/bskylog/scanner"
)

type Processor struct {
	scanner  scanner.Scanner
	consumer consumer.Consumer
}

func New(
	scanner scanner.Scanner,
	consumer consumer.Consumer,
) *Processor {
	return &Processor{
		scanner:  scanner,
		consumer: consumer,
	}
}

func (p *Processor) Proc(ctx context.Context) error {
	if err := p.scanner.Scan(ctx, "posts_with_replies", false, func(feed []*bsky.FeedDefs_FeedViewPost) error {
		for _, post := range feed {
			if err := p.consumer.Consume(ctx, post); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (p *Processor) Close() {
	p.consumer.Close()
}

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
)

type DailyJSONRecord struct {
	baseDir  string
	location *time.Location

	year    int
	month   time.Month
	day     int
	w       io.WriteCloser
	encoder *json.Encoder

	logger *slog.Logger
}

var _ Consumer = (*DailyJSONRecord)(nil)

type DailyJSONRecordOption func(c *DailyJSONRecord)

func SetDailyJSONRecordLogger(l *slog.Logger) DailyJSONRecordOption {
	return func(c *DailyJSONRecord) {
		if l == nil {
			c.logger = slog.Default()
		} else {
			c.logger = l
		}
	}
}

func SetDailyJSONRecordLocation(loc *time.Location) DailyJSONRecordOption {
	return func(c *DailyJSONRecord) {
		if loc == nil {
			c.location = time.UTC
		} else {
			c.location = loc
		}
	}
}

func NewDailyJSONRecord(baseDir string, opts ...DailyJSONRecordOption) *DailyJSONRecord {
	ret := &DailyJSONRecord{
		baseDir:  baseDir,
		location: time.UTC,
		logger:   slog.Default(),
	}
	for _, f := range opts {
		f(ret)
	}
	return ret
}

const (
	fileFlag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	filePerm = 0644
	dirPerm  = 0755
)

func (c *DailyJSONRecord) ensureStream(now time.Time) error {
	year, month, day := now.Date()
	if day == c.day && month == c.month && year == c.year {
		return nil
	}

	c.Close(nil)

	c.year = year
	c.month = month
	c.day = day

	y := fmt.Sprintf("%04d", c.year)
	m := fmt.Sprintf("%02d", int(c.month))
	d := fmt.Sprintf("%02d", c.day)

	file := filepath.Join(c.baseDir, y, m, d)
	f, err := os.OpenFile(file, fileFlag, filePerm)
	if err == nil {
		c.w = f
		c.encoder = json.NewEncoder(c.w)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Join(c.baseDir, y, m), dirPerm); err != nil {
		return err
	}

	f, err = os.OpenFile(file, fileFlag, filePerm)
	if err != nil {
		return err
	}
	c.w = f
	c.encoder = json.NewEncoder(c.w)

	return nil
}

func (c *DailyJSONRecord) Consume(ctx context.Context, post *bsky.FeedDefs_FeedViewPost) error {
	record, ok := post.Post.Record.Val.(*bsky.FeedPost)
	if !ok {
		c.logger.Warn("record is not post type", "cid", post.Post.Cid)

		// skip
		return nil
	}

	t, err := time.Parse(time.RFC3339Nano, record.CreatedAt)
	if err != nil {
		c.logger.Warn("time parse error",
			"cid", post.Post.Cid,
			"created_at", record.CreatedAt,
			"err", err,
		)

		// skip
		return nil
	}
	t = t.In(c.location)

	if err := c.ensureStream(t); err != nil {
		c.logger.Error("ensureStream() error",
			"baseDir", c.baseDir,
			"time", t,
		)
		return err
	}

	if err := c.encoder.Encode(post); err != nil {
		return err
	}

	c.logger.Info("Consume", "time", record.CreatedAt, "cid", post.Post.Cid)
	return nil
}

func (c *DailyJSONRecord) Close(_ context.Context) {
	if c.w != nil {
		c.w.Close()
	}
}

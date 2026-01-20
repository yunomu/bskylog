package consumer

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"

	"github.com/bluesky-social/indigo/api/bsky"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type TerminalValue struct {
	TimeStamp int64
	Cid       string
}

type DailyJSONRecordS3 struct {
	s3Client S3Client
	bucket   string
	baseDir  string
	location *time.Location

	terminalValue *TerminalValue

	year  int
	month time.Month
	day   int

	key      string
	buf      *bytes.Buffer
	encoder  *json.Encoder
	index    map[int]int
	indexKey string

	first     func(*TerminalValue)
	saveFirst func(int64, string)
	keyUpdate func(string)

	logger *slog.Logger
}

var _ Consumer = (*DailyJSONRecordS3)(nil)

type DailyJSONRecordS3Option func(c *DailyJSONRecordS3)

func SetDailyJSONRecordS3Logger(logger *slog.Logger) DailyJSONRecordS3Option {
	return func(c *DailyJSONRecordS3) {
		if logger == nil {
			c.logger = slog.Default()
		} else {
			c.logger = logger
		}
	}
}

func SetDailyJSONRecordS3TerminalValue(v *TerminalValue) DailyJSONRecordS3Option {
	return func(c *DailyJSONRecordS3) {
		c.terminalValue = v
	}
}

func SetDailyJSONRecordS3FirstValueFunc(f func(v *TerminalValue)) DailyJSONRecordS3Option {
	return func(c *DailyJSONRecordS3) {
		c.first = f
	}
}

func SetDailyJSONRecordS3KeyUpdateFunc(f func(string)) DailyJSONRecordS3Option {
	return func(c *DailyJSONRecordS3) {
		c.keyUpdate = f
	}
}

func NewDailyJSONRecordS3(
	s3Client S3Client,
	bucket string,
	baseDir string,
	location *time.Location,
	opts ...DailyJSONRecordS3Option,
) *DailyJSONRecordS3 {
	ret := &DailyJSONRecordS3{
		s3Client:  s3Client,
		bucket:    bucket,
		baseDir:   baseDir,
		location:  location,
		first:     func(*TerminalValue) {},
		keyUpdate: func(string) {},
		logger:    slog.Default(),
	}
	for _, f := range opts {
		f(ret)
	}
	ret.saveFirst = func(ts int64, cid string) {
		ret.first(&TerminalValue{
			TimeStamp: ts,
			Cid:       cid,
		})
		ret.saveFirst = func(int64, string) {}
	}
	return ret
}

func (c *DailyJSONRecordS3) ensureStream(ctx context.Context, now time.Time) error {
	year, month, day := now.Date()
	if day == c.day && month == c.month && year == c.year {
		return nil
	}

	if err := c.Close(ctx); err != nil {
		return err
	}

	c.year = year
	c.month = month
	c.day = day

	c.key = fmt.Sprintf("%s/%04d/%02d/%02d", c.baseDir, year, int(month), day)

	c.buf = &bytes.Buffer{}

	c.encoder = json.NewEncoder(c.buf)

	c.index = make(map[int]int)
	c.indexKey = fmt.Sprintf("%s/%04d/%02d/index", c.baseDir, year, int(month))
	out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.indexKey),
	})
	if err != nil {
		var opErr *smithy.OperationError
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &opErr) && errors.As(opErr.Err, &noSuchKey) {
			// do nothing
		} else {
			c.logger.Error("s3.GetObject(index)",
				"bucket", c.bucket,
				"key", c.indexKey,
			)
			return err
		}
	} else {
		defer out.Body.Close()

		r := csv.NewReader(out.Body)

		// skip header
		_, err := r.Read()
		if err != nil {
			c.logger.Error("CSV read error (skip header)")
			return err
		}

		for {
			fields, err := r.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				c.logger.Error("CSV read error")
				return err
			}
			if len(fields) == 0 {
				break
			}

			d, err := strconv.Atoi(fields[0])
			if err != nil {
				c.logger.Error("atoi index day field",
					"fields[0]", fields[0],
				)
				return err
			}
			cnt, err := strconv.Atoi(fields[1])
			if err != nil {
				c.logger.Error("atoi index count field",
					"fields[1]", fields[1],
				)
				return err
			}
			c.index[d] = cnt
		}

		c.keyUpdate(c.indexKey)
	}

	return nil
}

func (c *DailyJSONRecordS3) Consume(ctx context.Context, post *bsky.FeedDefs_FeedViewPost) error {
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

	if c.terminalValue != nil {
		if c.terminalValue.Cid == post.Post.Cid || c.terminalValue.TimeStamp >= t.Unix() {
			return nil
		}
	}

	if err := c.ensureStream(ctx, t); err != nil {
		c.logger.Error("ensureStream() error",
			"baseDir", c.baseDir,
			"time", t,
		)
		return err
	}

	if err := c.encoder.Encode(post); err != nil {
		return err
	}
	c.saveFirst(t.Unix(), post.Post.Cid)
	c.index[t.Day()]++

	c.logger.Info("Consume", "time", record.CreatedAt, "cid", post.Post.Cid)
	return nil
}

func (c *DailyJSONRecordS3) appendObject(ctx context.Context) error {
	s3Out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
	})
	if err != nil {
		var opErr *smithy.OperationError
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &opErr) && errors.As(opErr.Err, &noSuchKey) {
			// do nothing
			return nil
		} else {
			c.logger.Error("s3.GetObject",
				"bucket", c.bucket,
				"key", c.key,
			)
			return err
		}
	}
	defer s3Out.Body.Close()

	if _, err = c.buf.ReadFrom(s3Out.Body); err != nil {
		return err
	}

	c.keyUpdate(c.key)

	return nil
}

type entry struct {
	day   int
	count int
}

type entrySlice []*entry

func (e entrySlice) Len() int               { return len(e) }
func (e entrySlice) Less(i int, j int) bool { return e[i].day < e[j].day }
func (e entrySlice) Swap(i int, j int)      { e[i], e[j] = e[j], e[i] }

func (c *DailyJSONRecordS3) Close(ctx context.Context) error {
	if c.buf == nil {
		return nil
	}

	if err := c.appendObject(ctx); err != nil {
		c.logger.Error("appendObject",
			"bucket", c.bucket,
			"key", c.key,
		)
		return err
	}

	if _, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.key),
		Body:   c.buf,
	}); err != nil {
		c.logger.Error("s3.PutObject",
			"bucket", c.bucket,
			"key", c.key,
		)
		return err
	}

	var index []*entry
	for k, v := range c.index {
		index = append(index, &entry{day: k, count: v})
	}
	sort.Sort(entrySlice(index))

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"day", "count"}); err != nil {
		c.logger.Error("index csv write header")
		return err
	}
	for _, e := range index {
		if err := w.Write([]string{
			fmt.Sprintf("%02d", e.day),
			fmt.Sprintf("%d", e.count),
		}); err != nil {
			c.logger.Error("index csv write field error")
			return err
		}
	}
	w.Flush()

	if _, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(c.indexKey),
		Body:   &buf,
	}); err != nil {
		c.logger.Error("s3.PutObject(index)",
			"bucket", c.bucket,
			"key", c.indexKey,
		)
		return err
	}

	return nil
}

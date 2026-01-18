package consumer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
)

func TestTimeParse(t *testing.T) {
	str := "2023-07-27T23:53:47.210Z"
	loc := time.FixedZone("Local", 9*60*60)

	tm, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tm = tm.In(loc)

	t.Logf("time: %v", tm)
}

func TestRecord(t *testing.T) {
	s := `{"post":{"$type":"","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:testuser/testimage@jpeg","createdAt":"2023-07-27T23:53:47.210Z","did":"did:plc:testuser","displayName":"test","handle":"example.bsky.app","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"cid","indexedAt":"2026-01-01T01:02:03.100Z","likeCount":0,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-01-01T02:03:04.500Z","langs":["ja"],"text":"わたくし"},"replyCount":0,"repostCount":0,"uri":"at://did:plc:testuser/app.bsky.feed.post/postid","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}}}`

	var record bsky.FeedDefs_FeedViewPost
	if err := json.Unmarshal([]byte(s), &record); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	post, ok := record.Post.Record.Val.(*bsky.FeedPost)
	if !ok {
		t.Fatalf("post: %#v", record.Post.Record.Val)
	}

	t.Logf("post: %v", post)
}

func TestDailyJSONRecord_ensureStream(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	c := NewDailyJSONRecord(dir)
	defer c.Close(nil)

	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	if err := c.ensureStream(now); err != nil {
		t.Errorf("ensureStream: %v", err)
	}
}

func TestDailyJSONRecord_ensureStream_append(t *testing.T) {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(dir)

	dirname := fmt.Sprintf("%s/2026/01", dir)
	filename := fmt.Sprintf("%s/01", dirname)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := func() error {
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := fmt.Fprintln(f, "1"); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		t.Fatalf("Create test file: %v", err)
	}

	if err := func() error {
		c := NewDailyJSONRecord(dir)
		defer c.Close(nil)

		now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		if err := c.ensureStream(now); err != nil {
			return err
		}

		if _, err := fmt.Fprintln(c.w, "2"); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		t.Fatalf("Append: %v", err)
	}

	f, err := os.Open(filename)
	if err != nil {
		t.Fatalf("os.Open(%s): %v", filename, err)
	}

	var count int
	r := bufio.NewReader(f)
	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("ReadLine: %v", err)
		}

		t.Logf("line[%d]=%s", count, line)
		count++
	}

	if count != 2 {
		t.Errorf("count expected=2 actual=%d", count)
	}
}

package sqlite

import (
	"time"

	"gorm.io/gorm"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Record struct {
	Cid       string `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Text      string
	Timestamp int64

	Did    string
	Handle string
	Name   string

	ReplyParentHandle *string
	ReplyParentDid    *string
	ReplyParentName   *string

	Embed           string
	EmbedPostHandle *string
	EmbedPostDid    *string
	EmbedPostName   *string

	Key      string
	position int32
}

func ToRecord(key string, position int, post *bsky.FeedDefs_FeedViewPost) *Record {
	rec := &Record{
		Key:      key,
		position: int32(position),
		Embed:    "none",
	}

	if post == nil || post.Post == nil {
		return nil
	}

	rec.Cid = post.Post.Cid
	if post.Post.Author != nil {
		rec.Did = post.Post.Author.Did
		rec.Handle = post.Post.Author.Handle
		if post.Post.Author.DisplayName != nil {
			rec.Name = *post.Post.Author.DisplayName
		}
	}

	if feedPost, ok := post.Post.Record.Val.(*bsky.FeedPost); ok && feedPost != nil {
		rec.Text = feedPost.Text
		if createdAt, err := time.Parse(time.RFC3339, feedPost.CreatedAt); err == nil {
			rec.Timestamp = createdAt.UnixMicro()
		}
	}

	if post.Reply != nil && post.Reply.Parent != nil && post.Reply.Parent.FeedDefs_PostView != nil {
		p := post.Reply.Parent.FeedDefs_PostView
		if post.Reply.Parent.FeedDefs_PostView.Author != nil {
			rec.ReplyParentDid = &p.Author.Did
			rec.ReplyParentHandle = &p.Author.Handle
			if p.Author.DisplayName != nil {
				rec.ReplyParentName = p.Author.DisplayName
			}
		}
	}

	if post.Post.Embed != nil {
		if post.Post.Embed.EmbedImages_View != nil {
			rec.Embed = "image"
		} else if post.Post.Embed.EmbedExternal_View != nil {
			rec.Embed = "external"
		} else if post.Post.Embed.EmbedRecord_View != nil && post.Post.Embed.EmbedRecord_View.Record != nil {
			rec.Embed = "post"

			r := post.Post.Embed.EmbedRecord_View.Record.EmbedRecord_ViewRecord
			if r != nil {
				if r.Author != nil {
					rec.EmbedPostDid = &r.Author.Did
					rec.EmbedPostHandle = &r.Author.Handle
					rec.EmbedPostName = r.Author.DisplayName
				}
			}
		} else if post.Post.Embed.EmbedRecordWithMedia_View != nil {
			// For now, just record the media type
			if post.Post.Embed.EmbedRecordWithMedia_View.Media != nil {
				if post.Post.Embed.EmbedRecordWithMedia_View.Media.EmbedImages_View != nil {
					rec.Embed = "image"
				} else if post.Post.Embed.EmbedRecordWithMedia_View.Media.EmbedExternal_View != nil {
					rec.Embed = "external"
				}
			}
		}
	}

	return rec
}

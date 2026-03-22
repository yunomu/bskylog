package parquet

import (
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
)

type Record struct {
	Cid       string `parquet:"cid"`                              // cid
	Text      string `parquet:"text"`                             // record.text
	Timestamp int64  `parquet:"timestamp,timestamp(microsecond)"` // reocrd.createdAt

	Did    string `parquet:"did"`    // author.did
	Handle string `parquet:"handle"` // author.handle
	Name   string `parquet:"name"`   // author.displayName

	ReplyParentHandle *string `parquet:"replyParentHandle"` // reply.parent.author.handle
	ReplyParentDid    *string `parquet:"replyParentDid"`    // reply.parent.author.did
	ReplyParentName   *string `parquet:"replyParentName"`   // reply.parent.author.displayName

	Embed           string  `parquet:"embed"`           // value: none,image,video,external,post
	EmbedPostHandle *string `parquet:"embedPostHandle"` // embed.record.author.handle
	EmbedPostDid    *string `parquet:"embedPostDid"`    // embed.record.author.did
	EmbedPostName   *string `parquet:"embedPostName"`   // embed.record.author.displayName

	Key      string `parquet:"key"`      // S3 key
	position int32  `parquet:"position"` // position in the file
}

func toRecord(key string, position int, post *bsky.FeedDefs_FeedViewPost) *Record {
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

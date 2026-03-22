package parquet

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_toRecord(t *testing.T) {
	text := `{"post":{"$type":"app.bsky.feed.defs#feedViewPost","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreibz3kp63xwcclijfxmb7ddkvkaokmvswrij5ek7f6f4ph6r6lzxaa","indexedAt":"2026-03-08T11:50:21.093Z","likeCount":0,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T11:50:19.803Z","langs":["ja"],"text":"ノーマルタイプが強すぎるんだよ"},"replyCount":0,"repostCount":0,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgkblzv6gk2r","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}}}`

	var fvp bsky.FeedDefs_FeedViewPost
	if err := json.Unmarshal([]byte(text), &fvp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	rec := toRecord("key", 1, &fvp)

	ts, err := time.Parse(time.RFC3339, "2026-03-08T11:50:19.803Z")
	if err != nil {
		t.Fatalf("time parse error: %v", err)
	}

	expected := &Record{
		Cid:       "bafyreibz3kp63xwcclijfxmb7ddkvkaokmvswrij5ek7f6f4ph6r6lzxaa",
		Text:      "ノーマルタイプが強すぎるんだよ",
		Timestamp: ts.UnixMicro(),
		Did:       "did:plc:spfskpvcqvyicwe6hn75sr4d",
		Handle:    "wagahai.info",
		Name:      "灘",
		Embed:     "none",
		Key:       "key",
		position:  1,
	}

	if diff := cmp.Diff(expected, rec, cmpopts.IgnoreUnexported(Record{})); diff != "" {
		t.Errorf("toRecord mismatch (-want +got):\n%s", diff)
	}

	if rec.position != expected.position {
		t.Errorf("position mismatch: want %d, got %d", expected.position, rec.position)
	}
}

func Test_toRecord_embed(t *testing.T) {
	text := `{"post":{"$type":"app.bsky.feed.defs#feedViewPost","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreicvjbfra2ucxpubnnquqg4v67vx66v6o5gysvtjbl7vctxpvoocti","embed":{"$type":"app.bsky.embed.images#view","images":[{"alt":"ポケットモンスターファイアレッドのスロットで777が出ている画像","aspectRatio":{"height":720,"width":1280},"fullsize":"https://cdn.bsky.app/img/feed_fullsize/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreienx5bhmdnmqb3sbqr4o74ad5vfoiutzuaqcg7yandixjpn7ejwky","thumb":"https://cdn.bsky.app/img/feed_thumbnail/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreienx5bhmdnmqb3sbqr4o74ad5vfoiutzuaqcg7yandixjpn7ejwky"}]},"indexedAt":"2026-03-08T12:27:07.492Z","likeCount":0,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T12:27:05.144Z","embed":{"$type":"app.bsky.embed.images","images":[{"alt":"ポケットモンスターファイアレッドのスロットで777が出ている画像","aspectRatio":{"height":720,"width":1280},"image":{"$type":"blob","ref":{"$link":"bafkreienx5bhmdnmqb3sbqr4o74ad5vfoiutzuaqcg7yandixjpn7ejwky"},"mimeType":"image/jpeg","size":408365}}]},"langs":["ja"],"text":"ｳﾋｮｰ"},"replyCount":0,"repostCount":0,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgkdnr2zmc2d","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}}}`

	var fvp bsky.FeedDefs_FeedViewPost
	if err := json.Unmarshal([]byte(text), &fvp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	rec := toRecord("key", 1, &fvp)

	ts, err := time.Parse(time.RFC3339, "2026-03-08T12:27:05.144Z")
	if err != nil {
		t.Fatalf("time parse error: %v", err)
	}

	expected := &Record{
		Cid:       "bafyreicvjbfra2ucxpubnnquqg4v67vx66v6o5gysvtjbl7vctxpvoocti",
		Text:      "ｳﾋｮｰ",
		Timestamp: ts.UnixMicro(),
		Did:       "did:plc:spfskpvcqvyicwe6hn75sr4d",
		Handle:    "wagahai.info",
		Name:      "灘",
		Embed:     "image",
		Key:       "key",
		position:  1,
	}

	if diff := cmp.Diff(expected, rec, cmpopts.IgnoreUnexported(Record{})); diff != "" {
		t.Errorf("toRecord mismatch (-want +got):\n%s", diff)
	}

	if rec.position != expected.position {
		t.Errorf("position mismatch: want %d, got %d", expected.position, rec.position)
	}
}

func Test_toRecord_link(t *testing.T) {
	text := `{"post":{"$type":"app.bsky.feed.defs#feedViewPost","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreibjfxwfe3z6mmis5nwmb5p6u4k27ji35ammiz3mgwck3mnmmq6vrq","embed":{"$type":"app.bsky.embed.external#view","external":{"description":"YouTube video by 初星学園","thumb":"https://cdn.bsky.app/img/feed_thumbnail/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreihdvs3gz4n3q3mty2vq2ehvgo2kzqnomx5zzhdtit3jsfhj7gkoam","title":"初星学園 「キミとセミブルー」Official Music Video (HATSUBOSHI GAKUEN - Kimi to Semi Blue)","uri":"https://youtu.be/Z-LWjF5J6Mw?si=c6MZLWh7RmOb1vd-"}},"indexedAt":"2026-03-08T12:47:31.797Z","likeCount":0,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T12:47:26.689Z","embed":{"$type":"app.bsky.embed.external","external":{"description":"YouTube video by 初星学園","thumb":{"$type":"blob","ref":{"$link":"bafkreihdvs3gz4n3q3mty2vq2ehvgo2kzqnomx5zzhdtit3jsfhj7gkoam"},"mimeType":"image/jpeg","size":735940},"title":"初星学園 「キミとセミブルー」Official Music Video (HATSUBOSHI GAKUEN - Kimi to Semi Blue)","uri":"https://youtu.be/Z-LWjF5J6Mw?si=c6MZLWh7RmOb1vd-"}},"facets":[{"features":[{"$type":"app.bsky.richtext.facet#link","uri":"https://youtu.be/Z-LWjF5J6Mw?si=c6MZLWh7RmOb1vd-"}],"index":{"byteEnd":24,"byteStart":0}}],"langs":["ja"],"text":"youtu.be/Z-LWjF5J6Mw?..."},"replyCount":0,"repostCount":0,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgkes5zhrc2r","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}}}`

	var fvp bsky.FeedDefs_FeedViewPost
	if err := json.Unmarshal([]byte(text), &fvp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	rec := toRecord("key", 1, &fvp)

	ts, err := time.Parse(time.RFC3339, "2026-03-08T12:47:26.689Z")
	if err != nil {
		t.Fatalf("time parse error: %v", err)
	}

	expected := &Record{
		Cid:       "bafyreibjfxwfe3z6mmis5nwmb5p6u4k27ji35ammiz3mgwck3mnmmq6vrq",
		Text:      "youtu.be/Z-LWjF5J6Mw?...",
		Timestamp: ts.UnixMicro(),
		Did:       "did:plc:spfskpvcqvyicwe6hn75sr4d",
		Handle:    "wagahai.info",
		Name:      "灘",
		Embed:     "external",
		Key:       "key",
		position:  1,
	}

	if diff := cmp.Diff(expected, rec, cmpopts.IgnoreUnexported(Record{})); diff != "" {
		t.Errorf("toRecord mismatch (-want +got):\n%s", diff)
	}

	if rec.position != expected.position {
		t.Errorf("position mismatch: want %d, got %d", expected.position, rec.position)
	}
}

func Test_toRecord_reply(t *testing.T) {
	text := `{"post":{"$type":"","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreihlk47alkx6tsnhbiyvdygfheegqwekt34j5dd74asujwoxn5p24u","indexedAt":"2026-03-08T06:09:22.991Z","likeCount":1,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T06:09:21.988Z","langs":["ja"],"reply":{"parent":{"cid":"bafyreibxb7ufiknpw2hih4bws5ggxc5clwelr3m5moodrblh72gjzynuvy","uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgjoc32hrc2r"},"root":{"cid":"bafyreibxb7ufiknpw2hih4bws5ggxc5clwelr3m5moodrblh72gjzynuvy","uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgjoc32hrc2r"}},"text":"なんか作ったことがある言語は余裕で10以上あるけどよく使うとか影響を受けているというレベルになると6～7くらいしかないな意外と"},"replyCount":0,"repostCount":0,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgjokds7hs2r","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}},"reply":{"parent":{"$type":"app.bsky.feed.defs#postView","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreibxb7ufiknpw2hih4bws5ggxc5clwelr3m5moodrblh72gjzynuvy","indexedAt":"2026-03-08T06:04:44.798Z","likeCount":6,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T06:04:44.386Z","langs":["ja"],"text":"私を構成する9つのプログラミング言語"},"replyCount":1,"repostCount":2,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgjoc32hrc2r","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}},"root":{"$type":"app.bsky.feed.defs#postView","author":{"associated":{"activitySubscription":{"allowSubscriptions":"followers"},"chat":{"allowIncoming":"all"}},"avatar":"https://cdn.bsky.app/img/avatar/plain/did:plc:spfskpvcqvyicwe6hn75sr4d/bafkreibt4azlojbs3zorzjqayugz4c3mu6t7hsifgh6hr3bp5cbesahxe4","createdAt":"2023-07-29T23:53:47.210Z","did":"did:plc:spfskpvcqvyicwe6hn75sr4d","displayName":"灘","handle":"wagahai.info","viewer":{"blockedBy":false,"muted":false}},"bookmarkCount":0,"cid":"bafyreibxb7ufiknpw2hih4bws5ggxc5clwelr3m5moodrblh72gjzynuvy","indexedAt":"2026-03-08T06:04:44.798Z","likeCount":6,"quoteCount":0,"record":{"$type":"app.bsky.feed.post","createdAt":"2026-03-08T06:04:44.386Z","langs":["ja"],"text":"私を構成する9つのプログラミング言語"},"replyCount":1,"repostCount":2,"uri":"at://did:plc:spfskpvcqvyicwe6hn75sr4d/app.bsky.feed.post/3mgjoc32hrc2r","viewer":{"bookmarked":false,"embeddingDisabled":false,"threadMuted":false}}}}`

	var fvp bsky.FeedDefs_FeedViewPost
	if err := json.Unmarshal([]byte(text), &fvp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	rec := toRecord("key", 1, &fvp)

	ts, err := time.Parse(time.RFC3339, "2026-03-08T06:09:21.988Z")
	if err != nil {
		t.Fatalf("time parse error: %v", err)
	}

	rDid := "did:plc:spfskpvcqvyicwe6hn75sr4d"
	rHandle := "wagahai.info"
	rName := "灘"
	expected := &Record{
		Cid:               "bafyreihlk47alkx6tsnhbiyvdygfheegqwekt34j5dd74asujwoxn5p24u",
		Text:              "なんか作ったことがある言語は余裕で10以上あるけどよく使うとか影響を受けているというレベルになると6～7くらいしかないな意外と",
		Timestamp:         ts.UnixMicro(),
		Did:               "did:plc:spfskpvcqvyicwe6hn75sr4d",
		Handle:            "wagahai.info",
		Name:              "灘",
		ReplyParentDid:    &rDid,
		ReplyParentHandle: &rHandle,
		ReplyParentName:   &rName,
		Embed:             "none",
		Key:               "key",
		position:          1,
	}

	if diff := cmp.Diff(expected, rec, cmpopts.IgnoreUnexported(Record{})); diff != "" {
		t.Errorf("toRecord mismatch (-want +got):\n%s", diff)
	}

	if rec.position != expected.position {
		t.Errorf("position mismatch: want %d, got %d", expected.position, rec.position)
	}
}

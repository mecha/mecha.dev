package blog

import (
	"io"
	"strings"

	"github.com/gorilla/feeds"
)

func WriteFeed(w io.Writer, numItems, page int, format string) error {
	posts, err := GetPosts(numItems, (page-1)*numItems)
	if err != nil {
		return err
	}

	feed := BuildFeed(posts)

	switch strings.ToLower(format) {
	default:
		fallthrough
	case "rss":
		feed.WriteRss(w)
	case "atom":
		feed.WriteAtom(w)
	case "json":
		feed.WriteJSON(w)
	}

	return nil
}

func BuildFeed(posts []*Post) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       "mecha.dev",
		Description: "Posts from mecha's blog",
		Link:        &feeds.Link{Href: "https://mecha.dev/blog"},
		Author:      &feeds.Author{Name: "Miguel Muscat", Email: "mail@mecha.dev"},
		Items:       make([]*feeds.Item, 0),
	}

	for _, post := range posts {
		href := "https://mecha.dev/blog/" + post.Slug

		feed.Items = append(feed.Items, &feeds.Item{
			Id:          href,
			Title:       post.Title,
			Link:        &feeds.Link{Href: href},
			Description: post.Excerpt,
			Author:      feed.Author,
			Created:     post.Date,
		})
	}

	return feed
}

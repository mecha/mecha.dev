package blog

import (
	"html/template"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mecha/mecha.dev/md"
)

type Post struct {
	Slug    string
	Title   string
	Excerpt string
	Body    template.HTML
	Date    time.Time
	Public  bool
}

func ParsePostFile(filepath string) (*Post, error) {
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	post, err := ParsePost(raw)
	if err != nil {
		return nil, err
	}

	post.Slug = PostSlugFromFile(filepath)
	return post, nil
}

func PostSlugFromFile(filepath string) string {
	base := path.Base(filepath)
	ext := path.Ext(filepath)
	slug := base[:len(base)-len(ext)]
	return slug
}

func ParsePost(raw []byte) (*Post, error) {
	post := &Post{}

	doc, err := md.Parse(raw)
	if err != nil {
		return post, err
	}

	post.Body = doc.Body

	for name, value := range doc.Head {
		switch strings.ToLower(name) {
		case "title":
			post.Title = value
		case "excerpt":
			post.Excerpt = value
		case "public":
			post.Public = strings.ToLower(value) == "true"
		case "date":
			date, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, err
			}
			post.Date = date
		default:
			slog.Warn("blog: unknown post property", "property", name)
		}
	}

	return post, nil
}

package blog

import (
	"html/template"
	"io"
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
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	post, err := ParsePost(file)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func PostSlugFromFile(filepath string) string {
	base := path.Base(filepath)
	ext := path.Ext(filepath)
	slug := base[:len(base)-len(ext)]
	return slug
}

func ParsePost(reader io.Reader) (*Post, error) {
	post := &Post{}

	doc, err := md.Parse(reader)
	if err != nil {
		return post, err
	}

	post.Body = doc.Body

	for name, value := range doc.Head {
		switch strings.ToLower(name) {
		case "slug":
			post.Slug = value
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

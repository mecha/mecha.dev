package blog

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
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

func ParsePostFile(fsys fs.FS, filepath string) (*Post, error) {
	file, err := fsys.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open post file: %w", err)
	}

	post, err := ParsePost(file)
	if err != nil {
		return nil, err
	}

	if post.Slug == "" {
		base := path.Base(filepath)
		ext := path.Ext(filepath)
		post.Slug = base[:len(base)-len(ext)]
	}

	return post, nil
}

func ParsePost(reader io.Reader) (*Post, error) {
	post := &Post{}

	doc, err := md.Parse(reader)
	if err != nil {
		return post, fmt.Errorf("failed to parse post markdown: %w", err)
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

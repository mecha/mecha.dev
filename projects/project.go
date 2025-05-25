package projects

import (
	"html/template"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/mecha/mecha.dev/md"
)

type Project struct {
	Name  string
	Desc  string
	URL   string
	Repo  string
	Langs string
	Body  template.HTML
}

func Parse(reader io.Reader) (*Project, error) {
	doc, err := md.Parse(reader)
	if err != nil {
		return nil, err
	}

	project := &Project{
		Body: doc.Body,
	}

	for key, val := range doc.Head {
		switch strings.ToLower(key) {
		case "name":
			project.Name = val
		case "desc":
			project.Desc = val
		case "repo":
			project.Repo = val
		case "url":
			project.URL = val
		case "langs":
			project.Langs = val
		default:
			slog.Warn("projects: unknown property", "property", key)
		}
	}

	return project, nil
}

func ParseFile(filepath string) (*Project, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return Parse(file)
}

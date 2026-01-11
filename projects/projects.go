package projects

import (
	"html/template"
	"io"
	"io/fs"
	"iter"
	"log/slog"
	"maps"
	"os"
	"path"
	"strings"

	"github.com/mecha/mecha.dev/md"
)

var projects = map[string]*Project{}

type Project struct {
	ID    string
	Name  string
	Desc  string
	URL   string
	Repo  string
	Langs string
	Body  template.HTML
}

func GetAll() iter.Seq[*Project] {
	return maps.Values(projects)
}

func Delete(id string) bool {
	_, has := projects[id]
	if has {
		delete(projects, id)
	}
	return has
}

func LoadFromFs(fsys fs.FS) (int, error) {
	entries, err := fs.ReadDir(fsys, ".")

	if os.IsNotExist(err) {
		entries = []os.DirEntry{}
	} else if err != nil {
		return 0, err
	}

	num := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		_, err := LoadFromFile(fsys, name)
		if err != nil {
			return num, err
		}
		num++
	}

	slog.Info("projects: loaded projects from fs", slog.Int("num", num))
	return num, nil
}

func LoadFromFile(fsys fs.FS, filepath string) (*Project, error) {
	project, err := ParseFile(fsys, filepath)
	if err != nil {
		return nil, err
	}
	id := IDFromFilePath(filepath)
	projects[id] = project
	return project, nil
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

func ParseFile(fsys fs.FS, filepath string) (*Project, error) {
	file, err := fsys.Open(filepath)
	if err != nil {
		return nil, err
	}
	return Parse(file)
}

func IDFromFilePath(filepath string) string {
	base := path.Base(filepath)
	ext := path.Ext(filepath)
	return base[:len(base)-len(ext)]
}

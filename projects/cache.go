package projects

import (
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
)

var cache = map[string]*Project{}

func GetAll() map[string]*Project {
	return cache
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

func LoadIntoCache(id string, project *Project) {
	delete(cache, id)
	cache[id] = project
}

func LoadFromFile(fsys fs.FS, filepath string) (*Project, error) {
	project, err := ParseFile(filepath)
	if err != nil {
		return nil, err
	}
	id := path.Base(filepath)
	LoadIntoCache(id, project)
	return project, nil
}

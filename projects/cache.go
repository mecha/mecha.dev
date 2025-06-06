package projects

import (
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"
)

var cache = map[string]*Project{}

func GetAll() map[string]*Project {
	return cache
}

func LoadFromDir(dir string) error {
	slog.Info("projects: loading projects into cache from `" + dir + "`")
	entries, err := os.ReadDir(dir)

	if os.IsNotExist(err) {
		entries = []os.DirEntry{}
	} else if err != nil {
		return err
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

		_, err := LoadFromFile(dir + "/" + name)
		if err != nil {
			return err
		}
		num++
	}

	slog.Info("projects: loaded " + strconv.Itoa(num) + " projects")
	return nil
}

func LoadIntoCache(id string, project *Project) {
	delete(cache, id)
	cache[id] = project
}

func LoadFromFile(filepath string) (*Project, error) {
	project, err := ParseFile(filepath)
	if err != nil {
		return nil, err
	}
	id := path.Base(filepath)
	LoadIntoCache(id, project)
	return project, nil
}

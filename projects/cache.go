package projects

import (
	"log/slog"
	"os"
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
	if err != nil {
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

func LoadFromFile(path string) (*Project, error) {
	project, err := Parse(path)
	if err != nil {
		return nil, err
	}

	delete(cache, path)
	cache[path] = project
	return project, nil
}

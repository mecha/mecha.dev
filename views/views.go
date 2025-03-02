package views

import (
	t "html/template"
	"io"
	"log"
	"log/slog"
)

const baseTmplPath = "templates/base.gotmpl"

var cache = map[string]*t.Template{}

// Writes a view template to a writer with the given data. Failure will not be tolerated.
func Write(filename string, w io.Writer, data any) {
	err := getTemplate(filename).Execute(w, data)
	if err != nil {
		log.Fatal(err)
	}
}

// Retrieves the template object for a template file, consulting the cache first
// and populating it if missing for the given file.
func getTemplate(filepath string) *t.Template {
	if tmpl, isCached := cache[filepath]; isCached {
		return tmpl
	}
	tmpl := createTemplate(filepath)
	cache[filepath] = tmpl
	return tmpl
}

// Creates a template object from a template file
func createTemplate(filepath string) *t.Template {
	if filepath == baseTmplPath {
		return t.Must(t.ParseFiles(baseTmplPath)).Funcs(funcMap)
	} else {
		base := getTemplate(baseTmplPath)
		tmpl := t.Must(t.Must(base.Clone()).ParseFiles(filepath))
		return tmpl
	}
}

// Clears a single parsed view template from the cache
func ClearCache(filepath string) {
	if filepath == baseTmplPath {
		ClearAllCache()
	} else {
		delete(cache, filepath)
	}
}

// Clears the entire cache of parsed view templates
func ClearAllCache() {
	for key := range cache {
		delete(cache, key)
	}
	slog.Info("views: cleared all template caches")
}

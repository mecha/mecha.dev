package views

import (
	t "html/template"
	"io"
	"io/fs"
	"log"
	"log/slog"
)

var TemplateFS fs.FS
var cache = map[string]*t.Template{}

const baseTmplFilepath = "base.gotmpl"

// Writes a view template to a writer with the given data. Failure will not be tolerated.
func Write(filename string, w io.Writer, data any) {
	err := getTemplate(filename).Execute(w, data)
	if err != nil {
		log.Fatal("error writing view: " + err.Error())
	}
}

// Retrieves the template object for a template file, consulting the cache first
// and populating it if missing for the given file.
func getTemplate(filepath string) *t.Template {
	tmpl, isCached := cache[filepath]
	if !isCached {
		tmpl = createTemplate(filepath)
		cache[filepath] = tmpl
	}
	return tmpl
}

// Creates a template object from a template file
func createTemplate(filepath string) *t.Template {
	if filepath == baseTmplFilepath {
		return t.Must(t.New("base.gotmpl").Funcs(funcMap).ParseFS(TemplateFS, baseTmplFilepath))
	} else {
		base := getTemplate(baseTmplFilepath)
		tmpl := t.Must(t.Must(base.Clone()).ParseFS(TemplateFS, filepath))
		return tmpl
	}
}

// Clears a single parsed view template from the cache
func ClearCache(filepath string) {
	if filepath == baseTmplFilepath {
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

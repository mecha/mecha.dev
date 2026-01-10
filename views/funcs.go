package views

import (
	"html/template"
	"log/slog"
	"time"

	"github.com/mecha/mecha.dev/md"
)

// The functions available in view templates
var funcMap = template.FuncMap{
	"Now": time.Now,
	"IntRange": intRange,
	"MdFile": mdFile,
}

// Generates integers between start and end, inclusive and exclusive respectively.
func intRange(start, end int) (stream chan int) {
	stream = make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			stream <- i
		}
		close(stream)
	}()
	return
}

// Parses a markdown file and returns the HTML content, discarding front-matter.
// Uses the markdown cache.
func mdFile(path string) template.HTML {
	doc, err := md.ParseFileWithCache(path)
	if err != nil {
		slog.Error("error parsing markdown file: " + err.Error())
		return template.HTML("")
	}
	return doc.Body
}

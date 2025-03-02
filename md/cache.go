package md

import "log/slog"

// simple memory cache of parsed markdown documents
var cache = map[string]*ParsedDoc{}

// Parse a markdown file, consulting the cache first.
func ParseFileWithCache(filepath string) (*ParsedDoc, error) {
	doc, isCached := cache[filepath]
	if isCached {
		return doc, nil
	}

	doc, err := ParseFile(filepath)
	if err != nil {
		return nil, err
	}
	cache[filepath] = doc

	return doc, nil
}

// Clears a single entry from the cache.
func ClearCache(filepath string) bool {
	_, wasCached := cache[filepath]
	delete(cache, filepath)
	return wasCached
}

// Clears the entire cache.
func ClearAllCache() {
	for key := range cache {
		delete(cache, key)
	}
	slog.Info("md: cleared all parsed markdown caches")
}


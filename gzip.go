package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func gzipHandler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		accept := req.Header.Get("Accept-Encoding")
		if strings.Contains(accept, "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gw := gzip.NewWriter(w)
			defer gw.Close()
			w = GzipResponseWriter{ResponseWriter: w, Writer: gw}
		}

		handler.ServeHTTP(w, req)
	}
}

type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (gw GzipResponseWriter) Write(b []byte) (int, error) {
	if gw.Header().Get("Content-Type") == "" {
		gw.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return gw.Writer.Write(b)
}

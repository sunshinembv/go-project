package gzip

import (
	"compress/gzip"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") &&
		!strings.HasPrefix(contentType, "text/html") {
		return w.ResponseWriter.Write(data)
	}
	if w.writer == nil {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length")
		w.writer = gzip.NewWriter(w.ResponseWriter)
	}
	return w.writer.Write(data)

}

func (w *gzipResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *gzipResponseWriter) Close() error {
	if w.writer != nil {
		return w.writer.Close()
	}
	return nil
}

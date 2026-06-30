package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GunzipMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(
			strings.ToLower(ctx.GetHeader("Content-Encoding")),
			"gzip",
		) {
			ctx.Next()
			return
		}
		r, err := gzip.NewReader(ctx.Request.Body)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx.Request.Body = io.NopCloser(r)
		defer r.Close()
		ctx.Next()
	}
}

func GzipMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !strings.Contains(
			strings.ToLower(ctx.GetHeader("Accept-Encoding")),
			"gzip",
		) {
			ctx.Next()
			return
		}
		gzw := &gzipResponseWriter{
			ResponseWriter: ctx.Writer,
		}
		ctx.Writer = gzw
		defer gzw.Close()
		ctx.Next()
	}
}

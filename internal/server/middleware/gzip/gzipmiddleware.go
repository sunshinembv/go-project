package gzip

import (
	"compress/gzip"
	"fmt"
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
		defer func() {
			if err := r.Close(); err != nil {
				fmt.Printf("Failed to close resource: %v\n", err)
			}
		}()
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
		defer func() {
			if err := gzw.Close(); err != nil {
				fmt.Printf("Failed to close resource: %v\n", err)
			}
		}()
		ctx.Next()
	}
}

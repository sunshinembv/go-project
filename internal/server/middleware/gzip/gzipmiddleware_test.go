package gzip

import (
	"bytes"
	stdgzip "compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gzipData(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	writer := stdgzip.NewWriter(&buf)
	_, err := writer.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return buf.Bytes()
}

func gunzipData(t *testing.T, data []byte) []byte {
	t.Helper()
	reader, err := stdgzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer reader.Close()
	result, err := io.ReadAll(reader)
	require.NoError(t, err)
	return result
}

func TestGunzipMiddleware(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	type want struct {
		statusCode int
		body       string
	}

	type test struct {
		name     string
		encoding string
		body     []byte
		want     want
	}

	tests := []test{
		{
			name: "plain body",
			body: []byte("plain"),
			want: want{
				statusCode: http.StatusOK,
				body:       "plain",
			},
		},
		{
			name:     "gzip body",
			encoding: "gzip",
			body:     gzipData(t, []byte("compressed")),
			want: want{
				statusCode: http.StatusOK,
				body:       "compressed",
			},
		},
		{
			name:     "case insensitive",
			encoding: "application/GZIP",
			body:     gzipData(t, []byte("compressed")),
			want: want{
				statusCode: http.StatusOK,
				body:       "compressed",
			},
		},
		{
			name:     "invalid gzip",
			encoding: "gzip",
			body:     []byte("invalid"),
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(GunzipMiddleware())
			router.POST("/", func(ctx *gin.Context) {
				data, err := io.ReadAll(ctx.Request.Body)
				require.NoError(t, err)
				ctx.String(http.StatusOK, string(data))
			})

			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(tc.body))
			if tc.encoding != "" {
				req.Header.Set("Content-Encoding", tc.encoding)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, tc.want.statusCode, resp.Code)
			assert.Equal(t, tc.want.body, resp.Body.String())
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	type test struct {
		name        string
		accept      string
		contentType string
		body        string
		compressed  bool
	}

	tests := []test{
		{
			name:        "client does not accept gzip",
			contentType: "application/json",
			body:        `{"ok":true}`,
		},
		{
			name:        "json",
			accept:      "gzip",
			contentType: "application/json; charset=utf-8",
			body:        `{"ok":true}`,
			compressed:  true,
		},
		{
			name:        "html case insensitive header",
			accept:      "br, GZIP",
			contentType: "text/html",
			body:        "<p>ok</p>",
			compressed:  true,
		},
		{
			name:        "binary is not compressed",
			accept:      "gzip",
			contentType: "application/octet-stream",
			body:        "binary",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(GzipMiddleware())
			router.GET("/", func(ctx *gin.Context) {
				ctx.Header("Content-Type", tc.contentType)
				_, err := ctx.Writer.WriteString(tc.body)
				require.NoError(t, err)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.accept != "" {
				req.Header.Set("Accept-Encoding", tc.accept)
			}
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			require.Equal(t, http.StatusOK, resp.Code)
			if tc.compressed {
				assert.Equal(t, "gzip", resp.Header().Get("Content-Encoding"))
				assert.Contains(t, resp.Header().Values("Vary"), "Accept-Encoding")
				assert.Equal(t, tc.body, string(gunzipData(t, resp.Body.Bytes())))
				return
			}

			assert.Empty(t, resp.Header().Get("Content-Encoding"))
			assert.Equal(t, tc.body, resp.Body.String())
		})
	}
}

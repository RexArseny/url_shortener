package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Middleware struct {
	logger *zap.Logger
}

func NewMiddleware(logger *zap.Logger) Middleware {
	return Middleware{
		logger: logger,
	}
}

func (m *Middleware) Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		path := ctx.Request.URL.Path
		raw := ctx.Request.URL.RawQuery

		ctx.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		m.logger.Info("Request",
			zap.Int("code", ctx.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("method", ctx.Request.Method),
			zap.String("path", path),
			zap.Int("size", ctx.Writer.Size()))
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (m *Middleware) Compressor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				m.logger.Error("Can not create gzip reader", zap.Error(err))
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": http.StatusText(http.StatusInternalServerError)})
				ctx.Abort()
				return
			}
			defer func() {
				err = reader.Close()
				if err != nil {
					m.logger.Error("Can not close gzip reader", zap.Error(err))
				}
			}()
			ctx.Request.Body = reader
		}

		if strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") &&
			(strings.Contains(ctx.GetHeader("Content-Type"), "application/json") ||
				strings.Contains(ctx.GetHeader("Content-Type"), "text/html")) {
			writer := gzip.NewWriter(ctx.Writer)
			defer func() {
				err := writer.Close()
				if err != nil {
					m.logger.Error("Can not close gzip writer", zap.Error(err))
				}
			}()
			ctx.Writer.Header().Set("Content-Encoding", "gzip")
			ctx.Writer = &gzipWriter{ctx.Writer, writer}
		}

		ctx.Next()
	}
}

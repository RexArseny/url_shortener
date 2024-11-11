package middlewares

import (
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
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		m.logger.Info("Request",
			zap.Int("code", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("size", c.Writer.Size()))
	}
}

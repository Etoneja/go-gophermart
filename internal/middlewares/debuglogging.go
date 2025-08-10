package middlewares

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

func (m *Middlewares) DebugLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		writer := &bodyWriter{ResponseWriter: c.Writer}
		c.Writer = writer

		start := time.Now()
		c.Next()

		logEvent := m.logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", time.Since(start))

		if len(requestBody) > 0 {
			logEvent = logEvent.RawJSON("request_body", requestBody)
		}

		if len(writer.body) > 0 {
			logEvent = logEvent.RawJSON("response_body", writer.body)
		}

		logEvent.Msg("HTTP request")
	}
}

type bodyWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

package middleware

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/zukadong/zeus"
	"github.com/zukadong/zeus/internal/strconv"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const XHeader = "X-Trace-Id"

type bodyLogWriter struct {
	zeus.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestLogger() zeus.MiddlewareFunc {
	return func(next zeus.HandlerFunc) zeus.HandlerFunc {
		return func(ctx *zeus.Context) error {
			traceId := ctx.GetRequestHeader(XHeader)
			if traceId == "" {
				traceId = uuid.NewString()
				ctx.Request.Header.Set(XHeader, traceId)
			}

			var body = make(map[string]any)
			if ctx.Request.Method == http.MethodPost {
				_ = json.Unmarshal(ctx.Buf, &body)
			} else {
				body = ctx.QueryMap()
			}

			slog.Info("new request",
				slog.String("traceId", traceId),
				slog.String("requestAction", ctx.Action),
				slog.String("requestMethod", ctx.Request.Method),
				slog.String("requestProto", ctx.Request.Proto),
				slog.String("requestClientIp", ctx.ClientIP()),
				slog.Any("requestBody", body),
			)
			return next(ctx)
		}
	}
}

func ResponseLogger() zeus.MiddlewareFunc {
	return func(next zeus.HandlerFunc) zeus.HandlerFunc {
		return func(ctx *zeus.Context) error {
			bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
			ctx.Writer = bodyLogWriter
			start := time.Now()

			if err := next(ctx); err != nil {
				ctx.Zeus.ErrHandler(ctx, err)
			}

			body := make(map[string]any)
			str := strconv.Bytes2String(bodyLogWriter.body.Bytes())
			callback := ctx.GetQuery("callback")
			if callback != "" {
				str = strings.ReplaceAll(str, callback+"(", "")
				str = strings.ReplaceAll(str, ");", "")
			}

			_ = json.Unmarshal(strconv.String2Bytes(str), &body)
			cost := time.Now().Sub(start).String()
			code := ctx.Writer.Status()

			if code >= http.StatusBadRequest {
				slog.Error("new response",
					slog.String("traceId", ctx.GetRequestHeader(XHeader)),
					slog.String("costTime", cost),
					slog.Int("responseStatus", code),
					slog.Any("responseBody", body))
			} else {
				slog.Info("new response",
					slog.String("traceId", ctx.GetRequestHeader(XHeader)),
					slog.String("costTime", cost),
					slog.Int("responseStatus", code),
					slog.Any("responseBody", body))
			}

			return nil
		}
	}
}

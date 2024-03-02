package middleware

import (
	"fmt"
	"github.com/zukadong/zeus"
	"github.com/zukadong/zeus/internal/strconv"
	"log/slog"
	"net/http"
	"runtime"
)

const stackSize = 4 << 10 // 4KB

func Recovery(out any) zeus.MiddlewareFunc {
	return func(next zeus.HandlerFunc) zeus.HandlerFunc {
		return func(ctx *zeus.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					stack := make([]byte, stackSize)
					length := runtime.Stack(stack, false)

					slog.Error("handler crashed",
						slog.String("action", ctx.Action),
						slog.Any("error", err),
						slog.String("stack", strconv.Bytes2String(stack[:length])),
					)
					_ = ctx.JSON(http.StatusInternalServerError, out)
				}
			}()
			return next(ctx)
		}
	}
}

package middleware

import (
	"encoding/base64"
	"github.com/zukadong/zeus"
	"github.com/zukadong/zeus/internal/strconv"
	"net/http"
	"strings"
)

const basic = "basic"

func BasicAuth(user, password string, out any) zeus.MiddlewareFunc {
	return func(next zeus.HandlerFunc) zeus.HandlerFunc {
		return func(ctx *zeus.Context) error {
			auth := ctx.GetRequestHeader("Authorization")
			l := len(basic)
			if len(auth) > l+1 && strings.EqualFold(auth[:l], basic) {
				buf, err := base64.StdEncoding.DecodeString(auth[l+1:])
				if err != nil {
					return err
				}
				cred := strconv.Bytes2String(buf)

				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						if user == cred[:i] && password == cred[i+1:] {
							return next(ctx)
						}
						return ctx.JSON(http.StatusMethodNotAllowed, out)
					}
				}
			}
			return ctx.JSON(http.StatusUnauthorized, out)
		}
	}
}

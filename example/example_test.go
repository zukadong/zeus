package example

import (
	"fmt"
	"github.com/zukadong/zeus"
	"github.com/zukadong/zeus/middleware"
	"net/http"
	"testing"
)

var (
	panicOut = map[string]any{
		"Err": 50000,
		"Msg": "server internal error",
	}

	basicOut = map[string]any{
		"Err": 40003,
		"Msg": "no auth",
	}
)

func TestExample(t *testing.T) {
	z := zeus.New()
	z.Use(middleware.RequestLogger(), middleware.ResponseLogger(), middleware.Recovery(panicOut), middleware.BasicAuth("zuka", "test", basicOut))

	z.Register("exec", func(ctx *zeus.Context) error {
		fmt.Println("eee")

		//time.Sleep(6)
		//panic("panic")
		var T = struct {
			Action string `json:"action"`
			Test   int    `json:"test"`
		}{}

		fmt.Println(ctx.Bind(&T))
		fmt.Println(T)
		return ctx.JSON(http.StatusOK, map[string]any{"code": 0, "message": "succeed", "data": "123"})
	})

	er := z.Run(":9888")
	fmt.Println(er)
}

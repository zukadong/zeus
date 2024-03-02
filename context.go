package zeus

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/zukadong/zeus/binder"
	"github.com/zukadong/zeus/render"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const defaultMemory = 3 << 20 // 3 MB

type Context struct {
	Action  string
	Request *http.Request
	Writer  ResponseWriter
	Buf     []byte

	res   responseWriter
	Zeus  *Zeus
	query url.Values
}

func newContext(z *Zeus) *Context {
	return &Context{Zeus: z}
}

func (c *Context) reset() {
	c.Writer = &c.res
	c.query = nil
	c.Buf = nil
}

// Bind is used to bind the request to v, set the default and validate the data.
func (c *Context) Bind(v any) error {
	switch c.Request.Method {
	case http.MethodGet:
		return c.ShouldBindQuery(v)
	case http.MethodPost:
		return c.ShouldBindJSON(v)
	default:
		return fmt.Errorf("unsupported method[%s]", c.Request.Method)
	}
}

// ShouldBindJSON is a shortcut for c.ShouldBindWith(obj, binder.Json).
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWith(obj, binder.Json)
}

// ShouldBindQuery is a shortcut for c.ShouldBindWith(obj, binder.Query).
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWith(obj, binder.Query)
}

// ShouldBindWith binds the passed struct pointer using the specified binder engine. See the binder package.
func (c *Context) ShouldBindWith(obj any, b binder.Binder) error {
	return b.Bind(c.Request, obj)
}

func (c *Context) parseRequest() error {
	switch c.Request.Method {
	case http.MethodGet:
		c.Action = c.GetQuery("Action")

	case http.MethodPost:
		bufBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bufBody))
		c.Action = gjson.GetBytes(bufBody, "Action").String()
		c.Buf = bufBody

	default:
		return fmt.Errorf("unsupported method[%s]", c.Request.Method)
	}

	if c.Action == "" {
		return fmt.Errorf("action is empty, method[%s]", c.Request.Method)
	}
	return nil
}

// Query parses and returns the query of the request.
func (c *Context) Query() url.Values {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query
}

// QueryMap parses and returns the query of the request.
func (c *Context) QueryMap() map[string]any {
	m := c.Query()
	d := make(map[string]any)
	for k, v := range m {
		if len(v) > 0 {
			d[k] = v[0]
		}
	}
	return d
}

func (c *Context) ClientIP() string {
	clientIp := strings.TrimSpace(c.Request.RemoteAddr)

	if ip := strings.TrimSpace(strings.Split(c.GetRequestHeader("X-Forwarded-For"), ",")[0]); ip != "" {
		clientIp = ip
	} else if ip = strings.TrimSpace(c.GetRequestHeader("X-Real-IP")); ip != "" {
		clientIp = ip
	} else {
		clientIp, _, _ = net.SplitHostPort(clientIp)
	}

	if clientIp == "::1" {
		clientIp = "127.0.0.1"
	}
	return clientIp
}

// GetQuery is equal to c.Query().Get(key).
func (c *Context) GetQuery(key string) string { return c.Query().Get(key) }

// GetRequestHeader is equal to c.Request().Header.Get(key).
func (c *Context) GetRequestHeader(key string) string { return c.Request.Header.Get(key) }

// IsWebSocket reports whether HTTP connection is WebSocket or not.
func (c *Context) IsWebSocket() bool {
	if c.Request.Method == "GET" &&
		c.Request.Header.Get("Connection") == "Upgrade" &&
		c.Request.Header.Get("Upgrade") == "websocket" {
		return true
	}
	return false
}

// ContentLength return the length of the request body.
func (c *Context) ContentLength() int64 { return c.Request.ContentLength }

// ContentType returns the Content-Type of the request without the charset.
func (c *Context) ContentType() (ct string) {
	ct = c.Request.Header.Get("Content-Type")
	if index := strings.IndexByte(ct, ';'); index > 0 {
		ct = strings.TrimSpace(ct[:index])
	}
	return
}

// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) error {
	c.Writer.WriteHeader(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return nil
	}
	return r.Render(c.Writer)
}

// String sends the string text to the client with status code.
func (c *Context) String(code int, format string, values ...any) error {
	return c.Render(code, render.String{Format: format, Data: values})
}

// JSON serializes the given struct as JSON into the response body.
func (c *Context) JSON(code int, obj any) error {
	return c.Render(code, render.JSON{Data: obj})
}

// JSONP serializes the given struct as JSON into the response body.
// It adds padding to response body to request data from a server residing in a different domain than the client.
// It also sets the Content-Type as "application/javascript".
func (c *Context) JSONP(code int, obj any) error {
	callback := c.GetQuery("callback")
	if callback == "" {
		return c.Render(code, render.JSON{Data: obj})
	}
	return c.Render(code, render.JsonpJSON{Callback: callback, Data: obj})
}

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == http.StatusNoContent:
		return false
	case status == http.StatusNotModified:
		return false
	}
	return true
}

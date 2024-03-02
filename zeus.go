package zeus

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
)

// HandlerFunc is the handler of the service.
type HandlerFunc func(*Context) error

// MiddlewareFunc is the handler middleware.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Zeus struct {
	handlers    map[string]HandlerFunc
	middlewares []MiddlewareFunc

	pool sync.Pool
	lock sync.RWMutex
}

func New() *Zeus {
	z := new(Zeus)

	z.handlers = make(map[string]HandlerFunc)
	z.pool.New = func() any { return newContext(z) }

	return z
}

func (z *Zeus) Run(addr ...string) (err error) {
	address := resolveAddress(addr)
	fmt.Printf("Listening and serving HTTP on %s\n", address)
	err = http.ListenAndServe(address, z)
	return
}

// AcquireContext acquires a Context from the pool.
func (z *Zeus) AcquireContext(r *http.Request, w http.ResponseWriter) *Context {
	c := z.pool.Get().(*Context)
	c.Request = r
	c.res.reset(w)
	c.reset()
	return c
}

// ReleaseContext releases a Context into the pool.
func (z *Zeus) ReleaseContext(c *Context) {
	c.reset()
	z.pool.Put(c)
}

func (z *Zeus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := z.AcquireContext(r, w)

	if err := c.parseRequest(); err != nil {
		z.ErrHandler(c, err)
		z.ReleaseContext(c)
		return
	}

	h := z.handleRequest
	h = applyMiddleware(h, z.middlewares...)

	if err := h(c); err != nil {
		z.ErrHandler(c, err)
		z.ReleaseContext(c)
		return
	}

	z.ReleaseContext(c)
	return
}

func (z *Zeus) ErrHandler(c *Context, err error) {
	rsp := map[string]any{"code": 5000, "message": err.Error()}
	_ = c.JSON(http.StatusInternalServerError, rsp)
}

// Use registers the global middlewares that apply to all the services.
func (z *Zeus) Use(mws ...MiddlewareFunc) {
	z.middlewares = append(z.middlewares, mws...)
}

// Register registers a service with the name and the handler.
func (z *Zeus) Register(action string, handler HandlerFunc, mws ...MiddlewareFunc) {
	if action == "" {
		panic("Zeus.Register: the action must not be empty")
	}

	if handler == nil {
		panic("Zeus.Register: the handler must not be empty")
	}

	handler = applyMiddleware(handler, mws...)

	z.lock.Lock()
	z.handlers[action] = handler
	z.lock.Unlock()
}

func (z *Zeus) handleRequest(c *Context) error {
	if c.Action == "" {
		return errors.New("no action")
	}

	if h, ok := z.Handler(c.Action); ok {
		return h(c)
	} else {
		return fmt.Errorf("invalid action[%s]", c.Action)
	}
}

func (z *Zeus) Handler(name string) (HandlerFunc, bool) {
	z.lock.RLock()
	h, ok := z.handlers[name]
	z.lock.RUnlock()
	return h, ok
}

func applyMiddleware(h HandlerFunc, middlewares ...MiddlewareFunc) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); port != "" {
			fmt.Printf("Environment variable PORT=\"%s\"\n", port)
			return ":" + port
		}
		fmt.Printf("Environment variable PORT is undefined. Using port :8080 by default")
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too many parameters")
	}
}

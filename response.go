package zeus

import (
	"fmt"
	"io"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

var _ ResponseWriter = &responseWriter{}

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter

	Status() int

	Size() int

	WriteString(string) (int, error)

	Written() bool

	WriteHeaderNow()
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Size() int {
	return w.size
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			fmt.Printf("[WARNING] Headers were already written. Wanted to override status code %d with %d\n", w.status, code)
		}
		w.status = code
	}
}

func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

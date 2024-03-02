package render

import (
	"fmt"
	"github.com/zukadong/zeus/internal/strconv"
	"net/http"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []any
}

var plainContentType = []string{"text/plain; charset=utf-8"}

// Render (String) writes data with custom ContentType.
func (r String) Render(w http.ResponseWriter) error {
	writeContentType(w, plainContentType)

	if len(r.Data) > 0 {
		_, err := fmt.Fprintf(w, r.Format, r.Data...)
		return err
	}

	_, err := w.Write(strconv.String2Bytes(r.Format))
	return err
}

// WriteContentType (String) writes Plain ContentType.
func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}

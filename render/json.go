package render

import (
	"encoding/json"
	"github.com/zukadong/zeus/internal/strconv"
	"html/template"
	"net/http"
)

var (
	jsonContentType  = []string{"application/json; charset=utf-8"}
	jsonpContentType = []string{"application/javascript; charset=utf-8"}
)

// JSON contains the given interface object.
type JSON struct {
	Data any
}

// JsonpJSON contains the given interface object its callback.
type JsonpJSON struct {
	Callback string
	Data     any
}

// Render (JSON) writes data with custom ContentType.
func (r JSON) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	_, err = w.Write(jsonBytes)
	return err
}

// WriteContentType (JSON) writes JSON ContentType.
func (r JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

// Render (JsonpJSON) marshals the given interface object and writes it and its callback with custom ContentType.
func (r JsonpJSON) Render(w http.ResponseWriter) (err error) {
	r.WriteContentType(w)
	ret, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	if r.Callback == "" {
		_, err = w.Write(ret)
		return err
	}

	callback := template.JSEscapeString(r.Callback)
	if _, err = w.Write(strconv.String2Bytes(callback)); err != nil {
		return err
	}

	if _, err = w.Write(strconv.String2Bytes("(")); err != nil {
		return err
	}

	if _, err = w.Write(ret); err != nil {
		return err
	}

	if _, err = w.Write(strconv.String2Bytes(");")); err != nil {
		return err
	}

	return nil
}

// WriteContentType (JsonpJSON) writes Javascript ContentType.
func (r JsonpJSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonpContentType)
}

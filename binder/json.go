package binder

import (
	"encoding/json"
	"errors"
	"net/http"
)

type jsonBinding struct{}

func (jsonBinding) Name() string { return "json" }

func (jsonBinding) Bind(r *http.Request, obj any) error {
	if r == nil || r.Body == nil {
		return errors.New("invalid request")
	}
	return json.NewDecoder(r.Body).Decode(obj)
}

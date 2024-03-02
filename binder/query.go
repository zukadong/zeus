package binder

import "net/http"

type queryBinding struct{}

func (queryBinding) Name() string { return "query" }

func (queryBinding) Bind(r *http.Request, obj any) error {
	values := r.URL.Query()

	if err := BindURLValues(obj, values, "query"); err != nil {
		return err
	}

	return nil
}

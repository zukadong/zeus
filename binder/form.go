package binder

import "net/http"

type formBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(*http.Request, any) error {
	// todo implemented
	return nil
}

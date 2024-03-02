package binder

import "net/http"

const (
	MIMEApplicationJSON = "application/json"
	MIMEApplicationXML  = "application/xml"
	MIMEApplicationForm = "application/x-www-form-urlencoded"
	MIMEMultipartForm   = "multipart/form-data"
	MIMETextPLAIN       = "text/plain"
	MIMETextHTML        = "text/html"
	MIMETextXML         = "text/xml"
)

var (
	Query = queryBinding{}
	Json  = jsonBinding{}
	Form  = formBinding{}
)

type Binder interface {
	Name() string
	Bind(*http.Request, any) error
}

package wrappers

import (
	"net/http"
)

type HTTPError struct {
	Sessioner
	Err error
	Code int
}

func NewError(e error, c int) *HTTPError {
	return &HTTPError{
		Err: e,
		Code: c,
	}
}

func (e HTTPError) Error() string {
	return e.Err.Error()
}
func (e HTTPError) Status() int {
	return e.Code
}
func (e HTTPError) HTTPWrite(w http.ResponseWriter, r *http.Request) error {
	msg := Response{
		Status: "error",
		Message: e.Err.Error(),
	}
	return returnJson(msg, e.Code, w)
}

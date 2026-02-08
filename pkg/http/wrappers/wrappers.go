package wrappers

import (
	"fmt"
	"net/url"
	"net/http"
	"encoding/json"
)

type Error interface {
	IResponse
	Error() string
	Status() int
}

type IResponse interface {
	HTTPWrite(w http.ResponseWriter, r *http.Request) error
}

type HTTPError struct {
	Err error
	Code int
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

type Handler func(r *http.Request) IResponse

type Response struct {
	Status string `json:"status"`
	Code int `json:"-"`
	Message string `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
func (resp *Response) HTTPWrite(w http.ResponseWriter, r *http.Request) error {
	return returnJson(resp, resp.Code, w)
}

func Success(r any) IResponse {
	return &Response{
		Status: "success",
		Code: http.StatusOK,
		Data: r,
	}
}

type Redirect struct {
	Loc *url.URL
	Code int
}
func (rd *Redirect) HTTPWrite(w http.ResponseWriter, r *http.Request) error {
	code := rd.Code
	if code != http.StatusMovedPermanently || code != http.StatusFound ||
		code != http.StatusSeeOther {
		code = http.StatusFound
	}
	http.Redirect(w, r, rd.Loc.String(), code)
	return nil
}

func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(r)
		// return error if there's any
		if resp == nil {
			err := Response{
				Status: "error",
				Code: http.StatusInternalServerError,
				Message: "Handler did not return",
			}
			err.HTTPWrite(w, r)
			return
		}

		if err := resp.HTTPWrite(w, r); err != nil {
			msg := Response{
				Status: "error",
				Code: http.StatusInternalServerError,
				Message: fmt.Sprintf("Unable to formulate response: %v", err),
			}
			msg.HTTPWrite(w, r)
		}
	}
}

func returnJson(resp any, code int, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")

	if encdata, err := json.Marshal(resp); err == nil {
		w.WriteHeader(code)
		w.Write(encdata)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

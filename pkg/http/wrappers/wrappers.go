package wrappers

import (
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/gorilla/sessions"
)

type Error interface {
	IResponse
	Error() string
	Status() int
}

type IResponse interface {
	Session(s *sessions.Session)
	SaveSessions(r *http.Request, w http.ResponseWriter)
	HTTPWrite(w http.ResponseWriter, r *http.Request) error
}

type Sessioner struct {
	Sessions []*sessions.Session
}
func (sr *Sessioner) Session(s* sessions.Session) {
	if sr.Sessions == nil {
		sr.Sessions = make([]*sessions.Session, 0, 1)
	}
	sr.Sessions = append(sr.Sessions, s)
}
func (sr *Sessioner) SaveSessions(r *http.Request, w http.ResponseWriter) {
	if sr.Sessions != nil {
		for _, s := range sr.Sessions {
			sessions.Save(r, w, s)
		}
	}
}


type Handler func(r *http.Request) IResponse

type Response struct {
	Sessioner
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

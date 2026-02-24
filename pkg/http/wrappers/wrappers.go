package wrappers

import (
	"errors"
	"net/http"
	"encoding/json"

	"github.com/gorilla/sessions"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/log"
)

var (
	l log.Logger
)

func SetLogger(logger log.Logger) {
	l = logger
}

type Request struct {
	U *db.User
	R *http.Request
	S *sessions.Session
	L log.Logger
}

type Error interface {
	IResponse
	Error() string
	Status() int
}

type IResponse interface {
	Session(s *sessions.Session)
	SaveSessions(r *http.Request, w http.ResponseWriter) error
	HTTPWrite(w http.ResponseWriter, r *http.Request) error
}

type Sessioner struct {
	Sessions []*sessions.Session `json:"-"`
}
func (sr *Sessioner) Session(s* sessions.Session) {
	if sr.Sessions == nil {
		sr.Sessions = make([]*sessions.Session, 0, 1)
	}
	sr.Sessions = append(sr.Sessions, s)
}
func (sr *Sessioner) SaveSessions(r *http.Request, w http.ResponseWriter) error {
	var err error
	if sr.Sessions != nil {
		for _, s := range sr.Sessions {
			if serr := s.Save(r, w); serr != nil {
				err = errors.Join(err, serr)
			}
		}
	}
	return err
}

type Response struct {
	Sessioner
	Status string `json:"status"`
	Code int `json:"-"`
	Message string `json:"message,omitempty"`
	Data interface{} `json:"data,omitempty"`
}
func (resp *Response) HTTPWrite(w http.ResponseWriter, r *http.Request) error {
	resp.SaveSessions(r, w)
	return returnJson(resp, resp.Code, w)
}

func Success(r any) IResponse {
	return &Response{
		Status: "success",
		Code: http.StatusOK,
		Data: r,
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

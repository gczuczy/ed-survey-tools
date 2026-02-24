package wrappers

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
)

type HandlerFunc func(r *Request) IResponse

type APIHandler struct {
	Methods map[string]Method
}

type Method struct {
	Handler HandlerFunc
	Auth bool
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		Methods: make(map[string]Method),
	}
}

func (h *APIHandler) Get(f HandlerFunc) *APIHandler {
	h.Methods["GET"] = Method{Handler: f}
	return h
}

func (h *APIHandler) AuthGet(f HandlerFunc) *APIHandler {
	h.Methods["GET"] = Method{Handler: f, Auth: true}
	return h
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m, found := h.Methods[r.Method]
	if !found {
		l.Info().Str("URL", r.URL.String()).Str("method", r.Method).
			Msgf("Method not found")
		http.NotFoundHandler().ServeHTTP(w, r)
		return
	}
	var resp IResponse
	req := Request{
		R: r,
		L: l.With().Str("URL", r.URL.String()).Str("method", r.Method).Logger(),
	}

	if m.Auth {
		sess, err := sessions.Get(r)
		if err != nil || sess.IsNew {
			l.Info().Str("URL", r.URL.String()).Str("method", r.Method).
				Msgf("User not found")
			resp = NewError(errors.Join(err, fmt.Errorf("Unauthorized")),
				http.StatusUnauthorized)
			formResponse(resp, w, r)
			return
		}
		req.S = sess
		if u, ok := sess.Values["user"].(*db.User); ok {
			req.U = u
		}
	}

	resp = m.Handler(&req)
	l.Info().Str("URL", r.URL.String()).Str("method", r.Method).
		Msgf("Served")
	formResponse(resp, w, r)
}

func formResponse(resp IResponse, w http.ResponseWriter, r *http.Request) {
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

package wrappers

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
)

type HandlerFunc func(r *Request) IResponse
// to check authorization
type AuthFunc func(r *Request) bool

type APIHandler struct {
	Methods map[string]Method
}

type Method struct {
	Handler HandlerFunc
	Auth bool
	AuthFuncs []AuthFunc
}

func NewAPIHandler() *APIHandler {
	return &APIHandler{
		Methods: make(map[string]Method),
	}
}

func (h *APIHandler) Get(f HandlerFunc) *APIHandler {
	h.Methods["GET"] = Method{
		Handler: f,
		Auth: false,
	}
	return h
}

func (h *APIHandler) AuthGet(f HandlerFunc, afs ...AuthFunc) *APIHandler {
	h.Methods["GET"] = Method{
		Handler: f,
		Auth: true,
		AuthFuncs: afs,
	}
	return h
}

func (h *APIHandler) Post(f HandlerFunc) *APIHandler {
	h.Methods["POST"] = Method{
		Handler: f,
		Auth: false,
	}
	return h
}

func (h *APIHandler) AuthPost(f HandlerFunc, afs ...AuthFunc) *APIHandler {
	h.Methods["POST"] = Method{
		Handler: f,
		Auth: true,
		AuthFuncs: afs,
	}
	return h
}

func (h *APIHandler) Put(f HandlerFunc) *APIHandler {
	h.Methods["PUT"] = Method{
		Handler: f,
		Auth: false,
	}
	return h
}

func (h *APIHandler) AuthPut(f HandlerFunc, afs ...AuthFunc) *APIHandler {
	h.Methods["PUT"] = Method{
		Handler: f,
		Auth: true,
		AuthFuncs: afs,
	}
	return h
}

func (h *APIHandler) Patch(f HandlerFunc) *APIHandler {
	h.Methods["PATCH"] = Method{
		Handler: f,
		Auth: false,
	}
	return h
}

func (h *APIHandler) AuthPatch(f HandlerFunc, afs ...AuthFunc) *APIHandler {
	h.Methods["PATCH"] = Method{
		Handler: f,
		Auth: true,
		AuthFuncs: afs,
	}
	return h
}

func (h *APIHandler) Delete(f HandlerFunc) *APIHandler {
	h.Methods["DELETE"] = Method{
		Handler: f,
		Auth: false,
	}
	return h
}

func (h *APIHandler) AuthDelete(f HandlerFunc, afs ...AuthFunc) *APIHandler {
	h.Methods["DELETE"] = Method{
		Handler: f,
		Auth: true,
		AuthFuncs: afs,
	}
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
		Vars: mux.Vars(r),
	}

	if m.Auth {
		sess, err := sessions.Get(r)
		if err != nil || sess.IsNew {
			req.L.Info().Msgf("User not found")
			resp = NewError(errors.Join(err, fmt.Errorf("Unauthorized")),
				http.StatusUnauthorized)
			formResponse(resp, w, r)
			return
		}
		req.S = sess
		if u, ok := sess.Values["user"].(*db.User); ok {
			req.U = u
		}

		// and now check the AuthFuncs for authorizations
		if m.AuthFuncs != nil {
			for _, af := range m.AuthFuncs {
				if !af(&req) {
					req.L.Info().Msgf("Forbidden")
					resp = NewError(errors.Join(err, fmt.Errorf("Forbidden")),
						http.StatusForbidden)
					formResponse(resp, w, r)
					return
				}
			}
		}
	}

	resp = m.Handler(&req)
	req.L.Info().Msgf("Served")
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

func IsAdmin(r *Request) bool {
	if r.U == nil {
		return false
	}
	return r.U.IsAdmin || r.U.IsOwner
}

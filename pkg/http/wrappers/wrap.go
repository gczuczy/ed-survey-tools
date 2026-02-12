package wrappers

import (
	"fmt"
	"errors"
	"net/http"

	"github.com/gczuczy/ed-survey-tools/pkg/db"
	"github.com/gczuczy/ed-survey-tools/pkg/http/sessions"
)

func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := Request{
			R: r,
		}
		resp := h(&req)
		formResponse(resp, w, r)
	}
}

func AuthWrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp IResponse
		req := Request{
			R: r,
		}

		sess, err := sessions.Get(r)
		if err != nil || sess.IsNew {
			resp = NewError(errors.Join(err, fmt.Errorf("Unauthorized")),
				http.StatusUnauthorized)
		} else {
			req.S = sess
			if u, ok := sess.Values["user"].(*db.User); ok {
				req.U = u
			}
			resp = h(&req)
		}
		formResponse(resp, w, r)
	}
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

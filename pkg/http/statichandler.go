package http

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func newStaticBundleHandler(root string) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			rel := strings.TrimPrefix(r.URL.Path, "/static/")

			// Reject path traversal
			if strings.Contains(rel, "..") {
				http.Error(w, "Bad Request",
					http.StatusBadRequest)
				return
			}

			fpath := filepath.Join(
				root, filepath.FromSlash(rel),
			)

			f, err := os.Open(fpath)
			if err != nil {
				if os.IsNotExist(err) {
					http.NotFound(w, r)
					return
				}
				http.Error(w,
					"Internal Server Error",
					http.StatusInternalServerError)
				return
			}
			defer f.Close()

			fi, err := f.Stat()
			if err != nil || fi.IsDir() {
				http.NotFound(w, r)
				return
			}

			if strings.HasSuffix(fpath, ".json.gz") {
				w.Header().Set(
					"Content-Type", "application/json")
				w.Header().Set(
					"Content-Encoding", "gzip")
			}

			http.ServeContent(
				w, r, fi.Name(), fi.ModTime(), f)
		})
}

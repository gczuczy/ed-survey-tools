package http

import (
	"os"
	"fmt"
	"mime"
	"path"
	"io/fs"
	"bytes"
	"strings"
	_ "embed"
	"net/http"

	"github.com/quay/claircore/pkg/tarfs"
)

func init() {
	// Register all extensions used by the embedded frontend
	// explicitly. Some Linux systems lack entries in their system
	// MIME database. Without a known type, net/http falls back to
	// content-sniffing which requires io.Seeker — not implemented
	// by tarfs — and panics with HTTP 500 "seeker can't seek".
	mime.AddExtensionType(".css",   "text/css; charset=utf-8")
	mime.AddExtensionType(".eot",   "application/vnd.ms-fontobject")
	mime.AddExtensionType(".html",  "text/html; charset=utf-8")
	mime.AddExtensionType(".ico",   "image/x-icon")
	mime.AddExtensionType(".js",    "application/javascript")
	mime.AddExtensionType(".svg",   "image/svg+xml")
	mime.AddExtensionType(".ttf",   "font/ttf")
	mime.AddExtensionType(".woff",  "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
}

//go:embed webroot.tar
var spaTarball []byte


type SPAHandler struct {
	index string
	fsys *tarfs.FS
	fshandler http.Handler
}

func newSPAHandler() (*SPAHandler, error) {
	reader := bytes.NewReader(spaTarball)

	h := SPAHandler{
		index: fmt.Sprintf("index.html"),
	}

	var err error
	if h.fsys, err = tarfs.New(reader); err != nil {
		return nil, err
	}
	h.fshandler = http.FileServerFS(h.fsys)

	return &h, nil
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	requestPath := path.Clean(r.URL.Path)

	if strings.HasPrefix(requestPath, `/`) {
		requestPath = requestPath[1:]
	}

	if len(requestPath)==0 {
		requestPath = h.index
	}

	var st fs.FileInfo
	var err error
	if st, err = h.fsys.Stat(requestPath); os.IsNotExist(err) || st.IsDir() {
		requestPath = h.index
	} else if err != nil {
		fmt.Printf(" err: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeFileFS(w, r, h.fsys, requestPath)
}

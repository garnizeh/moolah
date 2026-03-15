// Package web holds the embedded static assets (CSS, JS, images) served
// by the web UI binary. The embed.FS is populated at compile time, so the
// production binary is fully self-contained with no external file system
// dependencies at runtime.
package web

import (
	"embed"
	"io/fs"
)

//go:embed static
var staticFiles embed.FS

// StaticFS is the sub-filesystem rooted at the "static" directory.
// Mount it at /static/ in the HTTP mux:
//
// mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(web.StaticFS)))
var StaticFS, _ = fs.Sub(staticFiles, "static") //nolint:errcheck // fs.Sub on a known valid path never errors

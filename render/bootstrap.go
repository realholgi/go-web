package render

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var embeddedStaticFS embed.FS

var sharedStaticFS = func() fs.FS {
	files, err := fs.Sub(embeddedStaticFS, "static")
	if err != nil {
		panic(err)
	}
	return files
}()

var staticHandler = http.FileServer(http.FS(sharedStaticFS))

// StaticHandler serves all assets embedded below render/static.
// Mount it below /static/shared/ with http.StripPrefix.
func StaticHandler() http.Handler {
	return staticHandler
}

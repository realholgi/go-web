package render

import (
	"embed"
	"errors"
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

// StaticHandler serves application assets from appFS and shared assets embedded
// below render/static. appFS must contain a static directory.
func StaticHandler(appFS fs.FS) http.Handler {
	appStaticFS, err := fs.Sub(appFS, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(overlayFS{appStaticFS, sharedStaticFS}))
}

type overlayFS []fs.FS

func (files overlayFS) Open(name string) (fs.File, error) {
	if name != "." && !fs.ValidPath(name) {
		return nil, fs.ErrInvalid
	}
	for _, filesystem := range files {
		file, err := filesystem.Open(name)
		if err == nil {
			return file, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}
	return nil, fs.ErrNotExist
}

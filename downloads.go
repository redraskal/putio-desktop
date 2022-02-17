package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/redraskal/putio-desktop/downloads"
	"github.com/sqweek/dialog"
)

func (b *App) downloadState(d downloads.Download) {
	log.Println("DEBUG DOWNLOAD STATE: ", d)
	// TODO
}

func (b *App) Queue(url string) {
	go func() {
		header, err := downloads.Head(url)
		if err != nil {
			// TODO: Send an alert to the client
			log.Println(err)
			return
		}
		ext := strings.TrimPrefix(filepath.Ext(header.FileName), ".")
		path, err := dialog.File().
			SetStartFile(header.FileName).
			Filter(fmt.Sprintf("%s (*.%s)", header.ContentType, ext), ext).
			Save()
		if err != nil {
			log.Println(err) // The client likely cancelled
			return
		}
		b.downloads.Queue(url, path)
	}()
}

func (b *App) ListDownloads() []downloads.Download {
	return b.downloads.Get()
}

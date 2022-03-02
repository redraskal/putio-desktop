package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/redraskal/putio-desktop/downloads"
	"github.com/sqweek/dialog"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (b *App) downloadState(d downloads.Download) {
	runtime.EventsEmit(b.ctx, "download_state", d)
}

func (b *App) Queue(url string) {
	go func() {
		header, err := downloads.Head(url)
		if err != nil {
			b.frontend.ExecJS("alert('Download failed, " + err.Error() + "');")
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

func (b *App) ShowDownload(id int) {
	d, err := b.downloads.Get(id)
	if err != nil {
		b.frontend.ExecJS("alert('File not found');")
		return
	}
	// TODO: Implementations for other platforms
	_ = exec.Command("explorer.exe", "/select,"+d.Path()).Run()
}

func (b *App) CountDownloading() int {
	return len(b.downloads.WithStatus(downloads.Downloading, downloads.Prefilling))
}

func (b *App) ListDownloads() []downloads.Download {
	return b.downloads.GetAll()
}

func (b *App) ClearCompleted() {
	b.downloads.ClearCompleted()
}

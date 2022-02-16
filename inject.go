package main

import (
	"log"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type js struct {
	runtime string
	ipc     string
	main    string
}

func (b *App) ReportPath(path string) {
	log.Printf("Current page reported: %s", path)
	// If true, still local- must fetch wailsjs before redirecting to put.io.
	b.currentPath = path
	if b.requiresRedirect() {
		runtime.EventsEmit(b.ctx, "report_file", "wails/runtime.js")
		runtime.EventsEmit(b.ctx, "report_file", "wails/ipc.js")
		runtime.EventsEmit(b.ctx, "report_file", "main.js")
	}
}

func (b *App) ReportFile(path string, content string) {
	log.Printf("File: %s, len: %d", path, len(content))
	switch path {
	case "wails/runtime.js":
		b.scripts.runtime = content
	case "wails/ipc.js":
		b.scripts.ipc = content
	case "main.js":
		b.scripts.main = content
	}
	if b.canRedirect() {
		runtime.EventsEmit(b.ctx, "redirect", "https://app.put.io")
		b.currentPath = "redirecting"
	}
}

func (b *App) injectJS() {
	b.frontend.ExecJS(b.scripts.runtime)
	b.frontend.ExecJS(b.scripts.ipc)
	b.frontend.ExecJS(b.scripts.main)
}

func (b *App) requiresRedirect() bool {
	return strings.HasSuffix(b.currentPath, "://wails/")
}

func (b *App) canRedirect() bool {
	return b.requiresRedirect() && b.scripts.runtime != "" && b.scripts.ipc != "" && b.scripts.main != ""
}

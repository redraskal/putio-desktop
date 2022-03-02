package main

import (
	"context"
	"log"
	"runtime"

	"github.com/redraskal/putio-desktop/downloads"
)

type App struct {
	ctx         context.Context
	frontend    frontend
	downloads   *downloads.Client
	currentPath string
	scripts     js
}

type frontend interface {
	ExecJS(js string)
}

func NewApp() *App {
	return &App{}
}

func (b *App) startup(ctx context.Context) {
	b.ctx = ctx
	if f := ctx.Value("frontend"); f != nil {
		b.frontend = f.(frontend)
	}
	if d, err := downloads.New(downloads.Options{
		Path:          ".",
		MaxConcurrent: 5,
		Splits:        5,
	}, b.downloadState); err == nil {
		b.downloads = d
		d.Run()
	}
}

func (b *App) domReady(ctx context.Context) {
	if !b.requiresRedirect() {
		b.injectJS()
	}
	// Enables window drag for macOS (this feature seems to break interactions on Windows).
	if runtime.GOOS == "darwin" {
		b.frontend.ExecJS(`document.querySelector('header').toggleAttribute("data-wails-drag");`)
	}
}

func (b *App) shutdown(ctx context.Context) {
	b.downloads.Shutdown()
}

func (b *App) Log(line string) {
	log.Printf("[JS] %s", line)
}

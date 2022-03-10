package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

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
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	workingDir := filepath.Dir(ex)
	if d, err := downloads.New(downloads.Options{
		Path:          workingDir,
		MaxConcurrent: 5,
		Splits:        5,
	}, b.downloadState); err == nil {
		b.downloads = d
		d.RunAndResume()
	}
}

func (b *App) domReady(ctx context.Context) {
	if !b.requiresRedirect() {
		b.injectJS()
	}
}

func (b *App) shutdown(ctx context.Context) {
	if err := b.downloads.Shutdown(); err != nil {
		log.Println(err)
	}
}

func (b *App) Log(line string) {
	log.Printf("[JS] %s", line)
}

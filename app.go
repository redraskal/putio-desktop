package main

import (
	"context"
	"log"
)

type App struct {
	ctx         context.Context
	frontend    frontend
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
}

func (b *App) domReady(ctx context.Context) {
	if !b.requiresRedirect() {
		b.injectJS()
	}
}

func (b *App) shutdown(ctx context.Context) {}

func (b *App) Log(line string) {
	log.Printf("[JS] %s", line)
}

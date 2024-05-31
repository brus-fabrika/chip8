package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Engine struct {
	Window   *sdl.Window
	Renderer *sdl.Renderer
}

func (e *Engine) Init() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}

	w, err := sdl.CreateWindow("SDL2 Test Window", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		SCREEN_WIDTH, SCREEN_HEIGHT, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL)
	if err != nil {
		return err
	}
	e.Window = w

	r, err := sdl.CreateRenderer(e.Window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return err
	}
	e.Renderer = r

	return nil
}

func (e *Engine) Destroy() {
	println("Destroying...")

	if e.Renderer != nil {
		e.Renderer.Destroy()
	}
	if e.Window != nil {
		e.Window.Destroy()
	}

	sdl.Quit()
}

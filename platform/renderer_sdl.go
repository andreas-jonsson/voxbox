// +build !js,!mobile

/*
Copyright (C) 2016 Andreas T Jonsson

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package platform

import (
	"image"
	"log"

	"github.com/goxjs/gl"
	"github.com/veandco/go-sdl2/sdl"
)

const fulscreenFlag = sdl.WINDOW_FULLSCREEN_DESKTOP //sdl.WINDOW_FULLSCREEN

type Config func(*sdlRenderer) error

func ConfigWithSize(w, h int) Config {
	return func(rnd *sdlRenderer) error {
		rnd.config.windowSize = image.Point{w, h}
		return nil
	}
}

func ConfigWithTitle(title string) Config {
	return func(rnd *sdlRenderer) error {
		rnd.config.windowTitle = title
		return nil
	}
}

func ConfigWithDiv(n int) Config {
	return func(rnd *sdlRenderer) error {
		rnd.config.resolutionDiv = n
		return nil
	}
}

func ConfigWithFulscreen(rnd *sdlRenderer) error {
	rnd.config.fulscreen = true
	return nil
}

func ConfigWithDebug(rnd *sdlRenderer) error {
	rnd.config.debug = true
	return nil
}

func ConfigWithNoVSync(rnd *sdlRenderer) error {
	rnd.config.novsync = true
	return nil
}

type sdlRenderer struct {
	window    *sdl.Window
	glContext sdl.GLContext

	config struct {
		windowTitle   string
		windowSize    image.Point
		resolutionDiv int
		debug, novsync,
		fulscreen bool
	}
}

func NewRenderer(configs ...Config) (*sdlRenderer, error) {
	var (
		err error
		rnd sdlRenderer
		dm  sdl.DisplayMode

		sdlFlags uint32 = sdl.WINDOW_SHOWN | sdl.WINDOW_OPENGL
	)

	for _, cfg := range configs {
		if err = cfg(&rnd); err != nil {
			return nil, err
		}
	}

	if err = sdl.GetDesktopDisplayMode(0, &dm); err != nil {
		return &rnd, err
	}

	cfg := &rnd.config
	if cfg.windowSize.X <= 0 {
		cfg.windowSize.X = int(dm.W)
	}
	if cfg.windowSize.Y <= 0 {
		cfg.windowSize.Y = int(dm.H)
	}

	if cfg.resolutionDiv > 0 {
		cfg.windowSize.X /= cfg.resolutionDiv
		cfg.windowSize.Y /= cfg.resolutionDiv
	}

	sdl.GL_SetAttribute(sdl.GL_RED_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_GREEN_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_BLUE_SIZE, 8)
	//sdl.GL_SetAttribute(sdl.GL_ALPHA_SIZE, 8)
	sdl.GL_SetAttribute(sdl.GL_DEPTH_SIZE, 24)
	//sdl.GL_SetAttribute(sdl.GL_STENCIL_SIZE, 8)

	//sdl.GL_SetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)

	sdl.GL_SetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GL_SetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 2)

	rnd.window, err = sdl.CreateWindow(cfg.windowTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, cfg.windowSize.X, cfg.windowSize.Y, sdlFlags)
	if err != nil {
		return &rnd, err
	}

	rnd.glContext, err = sdl.GL_CreateContext(rnd.window)
	if err != nil {
		return &rnd, err
	}

	gl.ContextWatcher.OnMakeCurrent(nil)
	if cfg.novsync {
		sdl.GL_SetSwapInterval(0)
	} else {
		sdl.GL_SetSwapInterval(1)
	}

	rnd.window.SetGrab(true)
	//sdl.ShowCursor(0)
	return &rnd, nil
}

func (rnd *sdlRenderer) ToggleFullscreen() {
	isFullscreen := (rnd.window.GetFlags() & fulscreenFlag) != 0
	if isFullscreen {
		rnd.window.SetFullscreen(0)
	} else {
		rnd.window.SetFullscreen(fulscreenFlag)
	}
}

func (rnd *sdlRenderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (rnd *sdlRenderer) Present() {
	sdl.GL_SwapWindow(rnd.window)
	if rnd.config.debug {
		checkGLError()
	}
}

func (rnd *sdlRenderer) Shutdown() {
	gl.ContextWatcher.OnDetach()
	sdl.GL_DeleteContext(rnd.glContext)
	rnd.window.Destroy()
}

func (rnd *sdlRenderer) SetWindowTitle(title string) {
	rnd.window.SetTitle(title)
}

func checkGLError() {
	if err := gl.GetError(); err != gl.NO_ERROR {
		log.Panicf("GL error: 0x%x\n", err)
	}
}

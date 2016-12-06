// +build mobile

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

package entry

import (
	"log"

	"github.com/andreas-jonsson/voxbox/platform"
	"github.com/andreas-jonsson/warp/game"
	"github.com/andreas-jonsson/warp/game/menu"
	"github.com/andreas-jonsson/warp/game/play"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
)

var (
	renderer     platform.Renderer
	gameInstance game.Game
)

func initialize() {
	err := platform.Init()
	if err != nil {
		log.Panicln(err)
	}
	renderer, err = platform.NewRenderer()
	if err != nil {
		log.Panicln(err)
	}

	states := map[string]game.GameState{
		"menu": menu.NewMenuState(),
		"play": play.NewPlayState(),
	}

	gameInstance, err = game.NewGame(states)
	if err != nil {
		log.Panicln(err)
	}

	var gctl game.GameControl = g
	if err := g.SwitchState("menu", gctl); err != nil {
		log.Panicln(err)
	}
}

func shutdown() {
	gameInstance.Shutdown()
	renderer.Shutdown()
	platform.Shutdown()
}

func Entry() {
	app.Main(func(a app.App) {
		var (
			glctx gl.Context
			sz    size.Event
		)

		paintDoneChan := make(chan struct{})

		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					initialize()
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					close(platform.PaintEventChan)
					close(platform.InputEventChan)
					shutdown()
					glctx = nil
				}
			case size.Event:
				sz = e
				select {
				case platform.InputEventChan <- e:
				default:
				}
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}

				if !g.Running() {
					// ????
				}

				renderer.Clear()

				if err := gameInstance.Update(); err != nil {
					log.Panicln(err)
				}

				if err := gameInstance.Render(); err != nil {
					log.Panicln(err)
				}

				renderer.Present()

				a.Publish()
				a.Send(paint.Event{})
			case touch.Event, key.Event:
				select {
				case platform.InputEventChan <- e:
				default:
				}
			}
		}
	})
}

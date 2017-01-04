// +build mobile

// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package entry

import (
	"log"

	"github.com/andreas-jonsson/voxbox/game"
	"github.com/andreas-jonsson/voxbox/game/menu"
	"github.com/andreas-jonsson/voxbox/game/play"
	"github.com/andreas-jonsson/voxbox/platform"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
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

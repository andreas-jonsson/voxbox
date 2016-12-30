// +build !mobile

// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package entry

import (
	"fmt"
	"log"

	"github.com/andreas-jonsson/voxbox/game"
	"github.com/andreas-jonsson/voxbox/game/menu"
	"github.com/andreas-jonsson/voxbox/game/play"
	"github.com/andreas-jonsson/voxbox/platform"
)

func Entry() {
	if err := platform.Init(); err != nil {
		log.Panicln(err)
	}
	defer platform.Shutdown()

	rnd, err := platform.NewRenderer(platform.ConfigWithDiv(2), platform.ConfigWithNoVSync) //, platform.ConfigWithDebug)
	if err != nil {
		log.Panicln(err)
	}
	defer rnd.Shutdown()

	platform.LogGLInfo()

	states := map[string]game.GameState{
		"menu": menu.NewMenuState(),
		"play": play.NewPlayState(),
	}

	g, err := game.NewGame(states)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Shutdown()

	var gctl game.GameControl = g
	if err := g.SwitchState("menu", gctl); err != nil {
		log.Panicln(err)
	}

	for g.Running() {
		rnd.Clear()

		if err := g.Update(); err != nil {
			log.Panicln(err)
		}

		_, _, fps := g.Timing()
		rnd.SetWindowTitle(fmt.Sprintf("Voxbox - %d fps", fps))

		if err := g.Render(); err != nil {
			log.Panicln(err)
		}

		rnd.Present()
	}
}

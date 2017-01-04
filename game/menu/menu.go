// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package menu

import (
	"github.com/andreas-jonsson/voxbox/game"
)

type menuState struct {
}

func NewMenuState() *menuState {
	return &menuState{}
}

func (s *menuState) Name() string {
	return "menu"
}

func (s *menuState) Enter(from game.GameState, args ...interface{}) error {
	return args[0].(game.GameControl).SwitchState("play", args[0])
}

func (s *menuState) Exit(to game.GameState) error {
	return nil
}

func (s *menuState) Update(gctl game.GameControl) error {
	gctl.PollAll()
	return nil
}

func (s *menuState) Render() error {
	return nil
}

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

package play

import (
	"image/color/palette"
	"log"
	"math"
	"time"

	"github.com/andreas-jonsson/voxbox/game"
	"github.com/andreas-jonsson/voxbox/game/player"
	"github.com/andreas-jonsson/voxbox/platform"
	"github.com/andreas-jonsson/voxbox/room"
	"github.com/andreas-jonsson/voxbox/view"
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/andreas-jonsson/warp/data"
	"github.com/goxjs/gl"
	"github.com/ungerik/go3d/mat4"
)

type playState struct {
	room   *room.Room
	view   *view.View
	player *player.Player
}

func NewPlayState() *playState {
	return &playState{}
}

func (s *playState) Name() string {
	return "play"
}

func loadRoom(r *room.Room, flags room.Flag) {
	voxPos := []voxel.Point{
		voxel.Pt(0, 0, 0),
		voxel.Pt(97, 0, 0),
		voxel.Pt(0, 0, 97),
		voxel.Pt(97, 0, 97),
	}

	for _, pos := range voxPos {
		if err := r.LoadVOXFile("test.vox", pos, flags); err != nil {
			log.Panicln(err)
		}
	}
}

func (s *playState) Enter(from game.GameState, args ...interface{}) error {
	s.room = room.NewRoom(voxel.Pt(256, 64, 256), 16*time.Millisecond)
	loadRoom(s.room, room.Flag(room.Attached))

	fp, err := data.FS.Open("test.vox")
	if err != nil {
		return err
	}
	defer fp.Close()

	img := voxel.NewPaletted(palette.Plan9, voxel.ZB)
	if err := vox.Decode(fp, img); err != nil {
		return err
	}

	v, err := view.NewView()
	if err != nil {
		return err
	}

	v.SetPalettes(img.Palette)
	s.view = v

	s.room.Start()

	s.player = player.NewPlayer(s.view)
	s.player.SetRoom(s.room)

	return nil
}

func (s *playState) Exit(to game.GameState) error {
	s.room.Destroy()
	s.view.Destroy()
	return nil
}

var anim = 0.0

func (s *playState) Update(gctl game.GameControl) error {
	dt, tick, _ := gctl.Timing()

	for ev := gctl.PollEvent(); ev != nil; ev = gctl.PollEvent() {
		switch t := ev.(type) {
		case *platform.KeyDownEvent:
			switch t.Key {
			case platform.KeyReturn:
				s.player.Die()
			case platform.KeyLeft:
				s.room.Send(func() {
					s.room.Clear()
					loadRoom(s.room, room.Flag(room.Falling))
				})
			}
		}
	}

	s.view.Clear(0)

	// ------------- update view ----------------

	anim += dt.Seconds() * 10

	<-s.room.Send(func() {
		//voxel.Blit(s.view, s.room, voxel.Pt(0, 0, int(anim)), s.room.Bounds())
		voxel.Blit(s.view, s.room, voxel.ZP, s.room.Bounds())
	})

	s.player.Render()

	// ------------------------------------------

	s.view.SetGLState()

	const (
		near        = 0.1
		far         = 10000
		angleOfView = 25
		aspectRatio = 16 / 9
	)

	scale := float32(math.Tan(angleOfView*0.5*math.Pi/180) * near)
	r := aspectRatio * scale
	l := -r
	t := scale
	b := -t

	var (
		viewMatrix,
		projMatrix mat4.T
	)

	/********/
	rot := float32(tick/time.Millisecond) * 0.0005
	viewMatrix.AssignEulerRotation(rot, math.Pi*0.25, 0)
	/********/

	viewMatrix.AssignEulerRotation(0, math.Pi*0.32, 0)
	//viewMatrix.TranslateY(-5)
	viewMatrix.TranslateZ(-320)

	projMatrix.AssignPerspectiveProjection(l, r, b, t, near, far)
	s.view.BuildBuffers(&projMatrix, &viewMatrix)

	gl.ClearColor(0.6, 0.6, 0.6, 1)

	return nil
}

func (s *playState) Render() error {
	return s.view.Render()
}

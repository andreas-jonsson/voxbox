// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package play

import (
	"fmt"
	"image/color/palette"
	"log"
	"math"
	"time"

	"github.com/andreas-jonsson/voxbox/data"
	"github.com/andreas-jonsson/voxbox/game"
	"github.com/andreas-jonsson/voxbox/game/player"
	"github.com/andreas-jonsson/voxbox/platform"
	"github.com/andreas-jonsson/voxbox/room"
	"github.com/andreas-jonsson/voxbox/view"
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/goxjs/gl"
	"github.com/ungerik/go3d/mat4"
)

type playState struct {
	room   room.Interface
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
		voxel.Pt(194, 0, 0),

		voxel.Pt(0, 0, 97),
		voxel.Pt(97, 0, 97),
		voxel.Pt(194, 0, 97),

		voxel.Pt(0, 0, 194),
		voxel.Pt(97, 0, 194),
		voxel.Pt(194, 0, 194),
	}

	for i, pos := range voxPos {
		if err := r.LoadVOXFile(fmt.Sprintf("test%d.vox", i%3), pos, flags); err != nil {
			log.Panicln(err)
		}
	}
	/*
		bounds := r.Bounds()
		for z := 0; z < bounds.Max.Z; z++ {
			for y := 0; y < bounds.Max.Y; y++ {
				for x := 0; x < bounds.Max.X; x++ {
					r.Set(x, y, z, 1)
				}
			}
		}
	*/
}

func (s *playState) Enter(from game.GameState, args ...interface{}) error {
	r := room.NewRoom(voxel.Pt(256, 64, 256), 16*time.Millisecond)
	loadRoom(r, room.Flag(room.Attached))

	fp, err := data.FS.Open("test0.vox")
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

	s.room = r.Start()

	s.player = player.NewPlayer(s.view)
	s.player.SetRoom(r)

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
				s.room.Clear()
				s.room.Send(func(r *room.Room) {
					loadRoom(r, room.Flag(room.Falling))
				})
			}
		}
	}

	s.view.Clear(0)

	// ------------- update view ----------------

	anim += dt.Seconds() * 10

	//<-s.room.Send(func() {
	//voxel.Blit(s.view, s.room, voxel.Pt(0, 0, int(anim)), s.room.Bounds())
	//})

	<-s.room.BlitToView(s.view, voxel.ZP, s.room.Bounds())

	s.player.Render()

	// ------------------------------------------

	s.view.SetGLState()

	const (
		near        = 0.1
		far         = 10000
		angleOfView = 45
		aspectRatio = 16 / 10
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

	viewMatrix.AssignEulerRotation(0, math.Pi*0.33, 0)
	viewMatrix.TranslateY(-20)
	viewMatrix.TranslateZ(-260)

	projMatrix.AssignPerspectiveProjection(l, r, b, t, near, far)
	s.view.BuildBuffers(&projMatrix, &viewMatrix)

	gl.ClearColor(0.6, 0.6, 0.6, 1)

	return nil
}

func (s *playState) Render() error {
	return s.view.Render()
}

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
	"image/color"
	"log"
	"math"
	"time"

	"github.com/andreas-jonsson/voxbox/game"
	"github.com/andreas-jonsson/voxbox/room"
	"github.com/andreas-jonsson/voxbox/view"
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/andreas-jonsson/warp/data"
	"github.com/goxjs/gl"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec3"
)

type voxelImage struct {
	pal  color.Palette
	size voxel.Point
	data []uint8
}

func (img *voxelImage) Bounds() voxel.Box {
	return voxel.Box{Min: voxel.ZP, Max: img.size}
}

func (img *voxelImage) SetBounds(b voxel.Box) {
	img.size = voxel.Pt(b.Max.X, b.Max.Z, b.Max.Y)
	sz := b.Max.X * b.Max.Y * b.Max.Z
	img.data = make([]uint8, sz)
}

func (img *voxelImage) SetPalette(pal color.Palette) {
	img.pal = pal
}

func (img *voxelImage) Set(x, y, z int, index uint8) {
	img.data[img.offset(x, z, y)] = index
}

func (img *voxelImage) Get(x, y, z int) uint8 {
	return img.data[img.offset(x, y, z)]
}

func (img *voxelImage) offset(x, y, z int) int {
	return z*img.size.X*img.size.Y + y*img.size.X + x
}

type playState struct {
	room *room.Room
	view *view.View
}

func NewPlayState() *playState {
	return &playState{}
}

func (s *playState) Name() string {
	return "play"
}

func (s *playState) Enter(from game.GameState, args ...interface{}) error {
	s.room = room.NewRoom(voxel.Pt(256, 64, 256), 160*time.Millisecond)
	if err := s.room.LoadVOXFile("test.vox", voxel.ZP, room.None); err != nil {
		log.Panicln(err)
	}
	s.room.FlagFloor(room.Indestructible)

	fp, err := data.FS.Open("test.vox")
	if err != nil {
		return err
	}
	defer fp.Close()

	var img voxelImage
	if err := vox.Decode(fp, &img); err != nil {
		return err
	}

	v, err := view.NewView()
	if err != nil {
		return err
	}

	v.SetPalettes(img.pal)
	s.view = v

	s.room.Start()

	return nil
}

func (s *playState) Exit(to game.GameState) error {
	s.room.Destroy()
	s.view.Destroy()
	return nil
}

var anim = 0.0

func (s *playState) Update(gctl game.GameControl) error {
	dt, _, _ := gctl.Timing()
	gctl.PollAll()

	s.view.Clear(0)

	// ------------- update view ----------------

	anim += dt.Seconds() * 10

	<-s.room.Send(func() {
		//voxel.Blit(s.view, s.room, voxel.Pt(0, 0, int(anim)), s.room.Bounds())
		voxel.Blit(s.view, s.room, voxel.ZP, s.room.Bounds())
	})

	// ------------------------------------------

	s.view.SetGLState()

	const (
		near        = 0.01
		far         = 1000.0
		angleOfView = 75
		aspectRatio = 16 / 9
	)

	scale := float32(math.Tan(angleOfView*0.5*math.Pi/180) * near)
	r := aspectRatio * scale
	l := -r
	t := scale
	b := -t

	var (
		m = mat4.Ident
		p mat4.T
	)

	rot := float32(90) //float32(tick/time.Millisecond) * 0.0005

	m.AssignEulerRotation(rot, 0, 0)
	m.TranslateY(-50)
	//m.TranslateX(10)

	p.AssignPerspectiveProjection(l, r, b, t, near, far)
	p.MultMatrix(&m)

	// Build buffers
	forward := vec3.T{0, 0, 1}
	q := m.Quaternion()
	q.Invert().RotateVec3(&forward)
	s.view.BuildBuffers(forward)

	voxelProgramID := s.view.ProgramID()
	mvp := gl.GetUniformLocation(voxelProgramID, "u_mvp")
	gl.UniformMatrix4fv(mvp, p.Slice())

	m.Invert()
	m.Transpose()

	minv := gl.GetUniformLocation(voxelProgramID, "u_model_inv")
	gl.UniformMatrix4fv(minv, m.Slice())

	gl.ClearColor(0.6, 0.6, 0.6, 1)

	return nil
}

func (s *playState) Render() error {
	return s.view.Render()
}

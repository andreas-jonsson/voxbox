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
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/andreas-jonsson/warp/data"
	"github.com/goxjs/gl"
	"github.com/goxjs/gl/glutil"
	"github.com/ungerik/go3d/mat4"
)

var (
	paletteData      []byte
	paletteTextureID gl.Texture

	vertexBuffer   []byte
	vertexBufferID gl.Buffer
	voxelProgramID gl.Program
)

type voxelImage struct {
	size voxel.Point
	data []uint8
}

func (img *voxelImage) SetBounds(b voxel.Box) {
	img.size = b.Max
	sz := b.Max.X * b.Max.Y * b.Max.Z
	img.data = make([]uint8, sz)

	vertexBuffer = make([]byte, 0, sz*12+sz*6)
}

func (img *voxelImage) SetPalette(pal color.Palette) {
	for _, c := range pal {
		r, g, b, _ := c.RGBA()
		paletteData = append(paletteData, byte(r))
		paletteData = append(paletteData, byte(g))
		paletteData = append(paletteData, byte(b))
	}
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
	img voxelImage
}

func NewPlayState() *playState {
	return &playState{}
}

func (s *playState) Name() string {
	return "play"
}

func (s *playState) Enter(from game.GameState, args ...interface{}) error {
	fp, err := data.FS.Open("test.vox")
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := vox.Decode(fp, &s.img); err != nil {
		return err
	}

	vertexBufferID = gl.CreateBuffer()
	voxelProgramID, err = glutil.CreateProgram(vertexShaderSrc, pixelShaderSrc)

	if len(paletteData) != 256*3 {
		log.Panicln("invalid palette")
	}

	paletteTextureID = gl.CreateTexture()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, paletteTextureID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, 16, 16, gl.RGB, gl.UNSIGNED_BYTE, paletteData)

	return err
}

func (s *playState) Exit(to game.GameState) error {
	return nil
}

func (s *playState) generateVoxel(c uint8, x, y, z int) {
	cubeVertices := []byte{
		// front
		0, 0, 1,
		1, 0, 1,
		1, 1, 1,
		0, 1, 1,
		// back
		0, 0, 0,
		1, 0, 0,
		1, 1, 0,
		0, 1, 0,
	}

	top := [...]int{
		1, 5, 6,
		6, 2, 1,
	}

	bottom := [...]int{
		4, 0, 3,
		3, 7, 4,
	}

	left := [...]int{
		4, 5, 1,
		1, 0, 4,
	}

	right := [...]int{
		3, 2, 6,
		6, 7, 3,
	}

	front := [...]int{
		0, 1, 2,
		2, 3, 0,
	}

	back := [...]int{
		7, 6, 5,
		5, 4, 7,
	}

	write := func(elements [6]int) {
		for i := 0; i < 6; i++ {
			offset := elements[i] * 3
			vertexBuffer = append(vertexBuffer, cubeVertices[offset]+byte(x))
			vertexBuffer = append(vertexBuffer, cubeVertices[offset+1]+byte(y))
			vertexBuffer = append(vertexBuffer, cubeVertices[offset+2]+byte(z))

			// Append color
			vertexBuffer = append(vertexBuffer, c)
		}
	}

	write(top)
	write(bottom)
	write(left)
	write(right)
	write(front)
	write(back)
}

func checkGLError() {
	if err := gl.GetError(); err != gl.NO_ERROR {
		log.Panicf("GL error: 0x%x\n", err)
	}
}

func (s *playState) Update(gctl game.GameControl) error {
	gctl.PollAll()

	img := &s.img
	vertexBuffer = vertexBuffer[:0]

	for z := 0; z < img.size.Z; z++ {
		for y := 0; y < img.size.Y; y++ {
			for x := 0; x < img.size.X; x++ {
				c := img.Get(x, y, z)
				if c > 0 {
					s.generateVoxel(c, x, y, z)
				}
			}
		}
	}

	gl.UseProgram(voxelProgramID)

	pal := gl.GetUniformLocation(voxelProgramID, "u_palette")
	gl.Uniform1i(pal, 0)

	pos := gl.GetAttribLocation(voxelProgramID, "a_position")

	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBufferID)
	gl.BufferData(gl.ARRAY_BUFFER, vertexBuffer, gl.STREAM_DRAW)

	gl.VertexAttribPointer(pos, 4, gl.UNSIGNED_BYTE, false, 0, 0)
	gl.EnableVertexAttribArray(pos)

	const (
		near        = 0.01
		far         = 1000
		angleOfView = 90
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

	_, tick, _ := gctl.Timing()
	rot := float32(tick/time.Millisecond) * 0.001

	m.AssignEulerRotation(rot, 0, 0)
	m.TranslateZ(-40)

	p.AssignPerspectiveProjection(l, r, b, t, 0.01, 100)

	p.MultMatrix(&m)

	mvp := gl.GetUniformLocation(voxelProgramID, "u_mvp")
	gl.UniformMatrix4fv(mvp, p.Slice())

	return nil
}

func (s *playState) Render() error {
	checkGLError()
	gl.Disable(gl.CULL_FACE)
	//gl.Disable(gl.DEPTH_TEST)

	//gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	checkGLError()
	gl.DrawArrays(gl.TRIANGLES, 0, len(vertexBuffer)/4)

	checkGLError()
	return nil
}

var vertexShaderSrc = `
	#version 120

	uniform mat4 u_mvp;
	attribute vec4 a_position;
	varying vec2 v_uv;

	void main()
	{
		v_uv = vec2(mod(a_position.w, 16) / 16, 1 - (a_position.w / 16) / 16);
	    gl_Position = u_mvp * vec4(a_position.xyz, 1);
	}
`

var pixelShaderSrc = `
	#version 120

	uniform sampler2D u_palette;
	varying vec2 v_uv;

	void main()
	{
		//gl_FragColor = vec4(1,0,0,1);

		gl_FragColor = texture2D(u_palette, v_uv);
	}
`

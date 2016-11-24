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
	"github.com/ungerik/go3d/vec3"
)

var (
	cubeVertices = [24]byte{
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

	leftIndices = [6]int{
		1, 5, 6,
		6, 2, 1,
	}

	rightIndices = [6]int{
		4, 0, 3,
		3, 7, 4,
	}

	topIndices = [6]int{
		4, 5, 1,
		1, 0, 4,
	}

	bottomIndices = [6]int{
		3, 2, 6,
		6, 7, 3,
	}

	frontIndices = [6]int{
		0, 1, 2,
		2, 3, 0,
	}

	backIndices = [6]int{
		7, 6, 5,
		5, 4, 7,
	}
)

var facesIndices = [][6]int{
	leftIndices,
	rightIndices,
	topIndices,
	bottomIndices,
	frontIndices,
	backIndices,
}

var facesNormals = []vec3.T{
	{1, 0, 0},
	{-1, 0, 0},
	{0, -1, 0},
	{0, 1, 0},
	{0, 0, 1},
	{0, 0, -1},
}

type faceName int

const (
	topFace faceName = iota
	bottomFace
	leftFace
	rightFace
	frontFace
	backFace
)

var (
	paletteData      []byte
	paletteTextureID gl.Texture
	voxelProgramID   gl.Program

	buffers [6]*faceBuffer
)

type faceBuffer struct {
	vertexBuffer   []byte
	vertexBufferID gl.Buffer
	indices        [6]int
	normal         vec3.T
	face           faceName
}

func newFaceBuffer(face faceName) *faceBuffer {
	return &faceBuffer{
		vertexBufferID: gl.CreateBuffer(),
		indices:        facesIndices[face],
		normal:         facesNormals[face],
		face:           face,
	}
}

func (b *faceBuffer) reset() {
	b.vertexBuffer = b.vertexBuffer[:0]
}

func (b *faceBuffer) append(x, y, z, color byte) {
	for i := 0; i < 6; i++ {
		index := b.indices[i] * 3
		b.vertexBuffer = append(b.vertexBuffer, cubeVertices[index]+x)
		b.vertexBuffer = append(b.vertexBuffer, cubeVertices[index+1]+y)
		b.vertexBuffer = append(b.vertexBuffer, cubeVertices[index+2]+z)
		b.vertexBuffer = append(b.vertexBuffer, color)
	}
}

func (b *faceBuffer) draw(location gl.Attrib) {
	if len(b.vertexBuffer) > 0 {
		gl.BindBuffer(gl.ARRAY_BUFFER, b.vertexBufferID)
		gl.BufferData(gl.ARRAY_BUFFER, b.vertexBuffer, gl.STREAM_DRAW)

		gl.VertexAttribPointer(location, 4, gl.UNSIGNED_BYTE, false, 0, 0)
		gl.EnableVertexAttribArray(location)

		gl.DrawArrays(gl.TRIANGLES, 0, len(b.vertexBuffer)/4)
	}
}

type voxelImage struct {
	size voxel.Point
	data []uint8
}

func (img *voxelImage) SetBounds(b voxel.Box) {
	img.size = voxel.Pt(b.Max.X, b.Max.Z, b.Max.Y)
	sz := b.Max.X * b.Max.Y * b.Max.Z
	img.data = make([]uint8, sz)
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

	for i := range buffers {
		buffers[i] = newFaceBuffer(faceName(i))
	}

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

	/*
		for z := 0; z < s.img.size.Z; z++ {
			for y := 0; y < s.img.size.Y; y++ {
				for x := 0; x < s.img.size.X; x++ {
					s.img.Set(x, y, z, byte(rand.Intn(15)+1))
				}
			}
		}
	*/

	return err
}

func (s *playState) Exit(to game.GameState) error {
	return nil
}

func isFaceExposed(img *voxelImage, x, y, z int, n vec3.T) bool {
	size := img.size
	x += int(n[0])
	y += int(n[1])
	z += int(n[2])

	if x < 0 || y < 0 || z < 0 || x >= size.X || y >= size.Y || z >= size.Z {
		return true
	}
	return img.Get(x, y, z) == 0
}

func (s *playState) Update(gctl game.GameControl) error {
	gctl.PollAll()

	for _, b := range buffers {
		b.reset()
	}

	img := &s.img
	for z := 0; z < img.size.Z; z++ {
		for y := 0; y < img.size.Y; y++ {
			for x := 0; x < img.size.X; x++ {
				c := img.Get(x, y, z)
				if c == 0 {
					continue
				}

				for _, b := range buffers {
					if isFaceExposed(img, x, y, z, b.normal) {
						b.append(byte(x), byte(y), byte(z), c)
					}
				}
			}
		}
	}

	gl.UseProgram(voxelProgramID)

	pal := gl.GetUniformLocation(voxelProgramID, "u_palette")
	gl.Uniform1i(pal, 0)

	const (
		near        = 0.01
		far         = 1000.0
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
	rot := float32(tick/time.Millisecond) * 0.0005

	m.AssignEulerRotation(rot, 0, 0)
	m.TranslateY(-50)

	p.AssignPerspectiveProjection(l, r, b, t, near, far)
	p.MultMatrix(&m)

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
	gl.Disable(gl.CULL_FACE)
	//gl.Disable(gl.DEPTH_TEST)
	//gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	pos := gl.GetAttribLocation(voxelProgramID, "a_position")
	normal := gl.GetUniformLocation(voxelProgramID, "u_normal")

	for _, b := range buffers {
		gl.Uniform3fv(normal, b.normal.Slice())
		b.draw(pos)
	}

	return nil
}

var vertexShaderSrc = `
	#version 120

	uniform mat4 u_mvp;
	uniform mat4 u_model_inv;
	uniform vec3 u_normal;

	attribute vec4 a_position;

	varying vec2 v_uv;
	varying vec3 v_light;

	void main()
	{
		int index = int(a_position.w);
		float x = mod(index, 16);
		float y = index / 16;

		v_uv = vec2(x / 16, 1 - y/16);

		// Lighting

		const vec3 lightColor = vec3(1,1,1);
		vec3 lightDir = normalize(vec3(1,1,1));

		const float ambientStrength = 0.5;
		vec3 ambient = ambientStrength * lightColor;

		vec3 normal = normalize(mat3(u_model_inv) * u_normal);

		float diff = max(dot(normal, lightDir), 0);
		vec3 diffuse = diff * lightColor;

		v_light = ambient + diffuse;
		gl_Position = u_mvp * vec4(a_position.xyz, 1);
	}
`

var pixelShaderSrc = `
	#version 120

	uniform sampler2D u_palette;

	varying vec2 v_uv;
	varying vec3 v_light;

	void main()
	{
		vec3 voxel_color = texture2D(u_palette, v_uv).xyz;
		voxel_color = vec3(1,0,0);

		gl_FragColor = vec4(voxel_color * v_light, 1);
	}
`

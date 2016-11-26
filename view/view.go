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

package view

import (
	"image/color"

	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/goxjs/gl"
	"github.com/goxjs/gl/glutil"
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

const (
	SizeX = 128
	SizeY = 64
	SizeZ = 128
)

func offset(x, y, z int) int {
	return z*SizeX*SizeY + y*SizeX + x
}

type View struct {
	paletteData      []byte
	paletteTextureID gl.Texture
	buffers          [6]*faceBuffer
	data             []uint8

	voxelProgramID gl.Program
	positionAttrib gl.Attrib
	normalUniform,
	palettesSampler gl.Uniform
}

func NewView() (*View, error) {
	v := &View{
		paletteTextureID: gl.CreateTexture(),
		data:             make([]uint8, SizeX*SizeY*SizeZ),
	}

	var err error
	v.voxelProgramID, err = glutil.CreateProgram(vertexShaderSrc, fragmentShaderSrc)
	if err != nil {
		return nil, err
	}

	v.positionAttrib = gl.GetAttribLocation(v.voxelProgramID, "a_position")
	v.normalUniform = gl.GetUniformLocation(v.voxelProgramID, "u_normal")
	v.palettesSampler = gl.GetUniformLocation(v.voxelProgramID, "u_palettes")

	for i := range v.buffers {
		v.buffers[i] = newFaceBuffer(faceName(i))
	}
	return v, nil
}

func (v *View) SetGLState() {
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)

	gl.UseProgram(v.voxelProgramID)
	gl.Uniform1i(v.palettesSampler, 0)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, v.paletteTextureID)
}

func (v *View) SetPalettes(palettes ...color.Palette) {
	v.paletteData = make([]byte, 256*256*3)

	for i, pal := range palettes {
		for j, c := range pal {
			r, g, b, _ := c.RGBA()
			idx := i*256 + j*3
			v.paletteData[idx] = byte(r)
			v.paletteData[idx+1] = byte(g)
			v.paletteData[idx+2] = byte(b)
		}
	}

	v.SetGLState()
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, 256, 256, gl.RGB, gl.UNSIGNED_BYTE, v.paletteData)
}

func (v *View) Bounds() voxel.Box {
	return voxel.Bx(0, 0, 0, SizeX, SizeY, SizeZ)
}

func (v *View) Set(x, y, z int, index uint8) {
	v.data[offset(x, y, z)] = index
}

func (v *View) Get(x, y, z int) uint8 {
	return v.data[offset(x, y, z)]
}

func (v *View) BuildBuffers(forward vec3.T) {
	var visibleBuffers []*faceBuffer

	for _, b := range v.buffers {
		b.reset()
		//if vec3.Dot(&b.normal, &forward) < 0 {
		visibleBuffers = append(visibleBuffers, b)
		//}
	}

	for z := 0; z < SizeZ; z++ {
		for y := 0; y < SizeY; y++ {
			for x := 0; x < SizeX; x++ {
				c := v.Get(x, y, z)
				if c == 0 {
					continue
				}

				for _, b := range visibleBuffers {
					if v.isFaceExposed(x, y, z, b.normal) {
						b.append(byte(x), byte(y), byte(z), c)
					}
				}
			}
		}
	}
}

func (v *View) isFaceExposed(x, y, z int, n vec3.T) bool {
	x += int(n[0])
	y += int(n[1])
	z += int(n[2])

	if x < 0 || y < 0 || z < 0 || x >= SizeX || y >= SizeY || z >= SizeZ {
		return true
	}
	return v.Get(x, y, z) == 0
}

func (v *View) Clear(c byte) {
	for z := 0; z < SizeZ; z++ {
		for y := 0; y < SizeY; y++ {
			for x := 0; x < SizeX; x++ {
				v.data[offset(x, y, z)] = c
			}
		}
	}
}

func (v *View) Render() error {
	for _, b := range v.buffers {
		gl.Uniform3fv(v.normalUniform, b.normal.Slice())
		b.draw(v.positionAttrib)
	}
	return nil
}

//TODO Remove the need for this.
func (v *View) ProgramID() gl.Program {
	return v.voxelProgramID
}

var vertexShaderSrc = `
	#version 120

	uniform mat4 u_mvp;
	uniform mat4 u_model_inv;
	uniform vec3 u_normal;

	attribute vec4 a_position;

	varying float v_color_index;
	varying vec3 v_light;

	void main()
	{
		v_color_index = a_position.w / 255;

		// Lighting

		const vec3 lightColor = vec3(1,1,1);
		vec3 lightDir = normalize(vec3(1,1,1));

		const float ambientStrength = 0.8;
		vec3 ambient = ambientStrength * lightColor;

		//vec3 normal = normalize(mat3(u_model_inv) * u_normal);
		vec3 normal = u_normal;

		float diff = max(dot(normal, lightDir), 0);
		vec3 diffuse = diff * lightColor;

		v_light = ambient + diffuse;
		gl_Position = u_mvp * vec4(a_position.xyz, 1);
	}
`

var fragmentShaderSrc = `
	#version 120

	uniform sampler2D u_palettes;

	varying float v_color_index;
	varying vec3 v_light;

	void main()
	{
		vec3 voxel_color = texture2D(u_palettes, vec2(v_color_index, 0)).xyz;
		gl_FragColor = vec4(voxel_color * v_light, 1);
	}
`
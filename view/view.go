// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package view

import (
	"image/color"
	"math"
	"sync"

	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/barnex/fmath"
	"github.com/goxjs/gl"
	"github.com/goxjs/gl/glutil"
	"github.com/ungerik/go3d/mat4"
	"github.com/ungerik/go3d/vec3"
)

const (
	SizeX = 160
	SizeY = 80
	SizeZ = 160
)

const (
	threadedBufferBuilds = false
	cullBackface         = true
	cullAngel            = 60
	nBuffers             = 1
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
	vertexBuffer    []byte
	vertexBufferIDs [nBuffers]gl.Buffer
	bufferCount     uint32
	indices         [6]int
	normal          vec3.T
	face            faceName
}

func newFaceBuffer(face faceName) *faceBuffer {
	var buffers [nBuffers]gl.Buffer
	for i := range buffers {
		buffers[i] = gl.CreateBuffer()
	}

	return &faceBuffer{
		vertexBufferIDs: buffers,
		indices:         facesIndices[face],
		normal:          facesNormals[face],
		face:            face,
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
		gl.BindBuffer(gl.ARRAY_BUFFER, b.vertexBufferIDs[b.bufferCount%nBuffers])
		gl.BufferData(gl.ARRAY_BUFFER, b.vertexBuffer, gl.STREAM_DRAW)

		gl.VertexAttribPointer(location, 4, gl.UNSIGNED_BYTE, false, 0, 0)
		gl.EnableVertexAttribArray(location)

		gl.DrawArrays(gl.TRIANGLES, 0, len(b.vertexBuffer)/4)
		b.bufferCount++
	}
}

func offset(x, y, z int) int {
	return z*SizeX*SizeY + y*SizeX + x
}

type View struct {
	paletteData      []byte
	paletteTextureID gl.Texture
	buffers          [6]*faceBuffer
	data             []uint8

	mvpMatrix,
	modelMatrix,
	viewMatrix mat4.T

	voxelProgramID gl.Program
	positionAttrib gl.Attrib
	normalUniform,
	palettesSampler gl.Uniform
}

func NewView() (*View, error) {
	v := &View{
		paletteTextureID: gl.CreateTexture(),
		modelMatrix:      mat4.Ident,
		data:             make([]uint8, SizeX*SizeY*SizeZ),
	}

	m := &v.modelMatrix
	m.TranslateX(-SizeX / 2)
	m.TranslateZ(-SizeZ / 2)

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

func (v *View) Destroy() {
	gl.DeleteProgram(v.voxelProgramID)
	gl.DeleteTexture(v.paletteTextureID)

	for i := range v.buffers {
		for _, id := range v.buffers[i].vertexBufferIDs {
			gl.DeleteBuffer(id)
		}
	}
}

func (v *View) Data() []uint8 {
	return v.data
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

func (v *View) BuildBuffers(proj, view *mat4.T) {
	v.viewMatrix = *view

	// Calculate MVP matrix.
	modelViewMatrix := v.viewMatrix
	modelViewMatrix.MultMatrix(&v.modelMatrix)
	v.mvpMatrix = *proj
	v.mvpMatrix.MultMatrix(&modelViewMatrix)

	m := modelViewMatrix.Array()
	forward := vec3.T{-m[2], -m[6], -m[10]}

	var visibleBuffers []*faceBuffer
	for _, b := range v.buffers {
		b.reset()

		angel := fmath.Acos(vec3.Dot(&b.normal, &forward)) / math.Pi * 180
		if !cullBackface || angel > cullAngel {
			visibleBuffers = append(visibleBuffers, b)
		}
	}

	//log.Println("Number of visible faces:", len(visibleBuffers))

	if threadedBufferBuilds {
		var wg sync.WaitGroup
		wg.Add(len(visibleBuffers))

		for _, b := range visibleBuffers {
			go func(b *faceBuffer) {
				for z := 0; z < SizeZ; z++ {
					for y := 0; y < SizeY; y++ {
						for x := 0; x < SizeX; x++ {
							c := v.Get(x, y, z)
							if c == 0 {
								continue
							}

							if v.isFaceExposed(x, y, z, b.normal) {
								b.append(byte(x), byte(y), byte(z), c)
							}
						}
					}
				}
				wg.Done()
			}(b)
		}

		wg.Wait()
	} else {
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
	mvp := gl.GetUniformLocation(v.voxelProgramID, "u_mvp")

	m := v.mvpMatrix
	gl.UniformMatrix4fv(mvp, m.Slice())

	m = v.viewMatrix
	m.MultMatrix(&v.modelMatrix)

	//m.Invert()
	//m.Transpose()

	mv := gl.GetUniformLocation(v.voxelProgramID, "u_mv")
	gl.UniformMatrix4fv(mv, m.Slice())

	for _, b := range v.buffers {
		gl.Uniform3fv(v.normalUniform, b.normal.Slice())
		b.draw(v.positionAttrib)
	}
	return nil
}

var vertexShaderSrc = `
	#version 120

	uniform mat4 u_mvp;
	uniform mat4 u_mv;
	uniform vec3 u_normal;

	attribute vec4 a_position;

	varying float v_color_index;
	varying vec3 v_light;

	void main()
	{
		v_color_index = a_position.w / 255;

		// Lighting

		const vec3 lightColor = vec3(1,1,1);
		vec3 lightDir = normalize(vec3(-1,1,-1));

		const float ambientStrength = 0.8;
		vec3 ambient = ambientStrength * lightColor;

		//vec3 normal = normalize(mat3(u_mv) * u_normal);
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

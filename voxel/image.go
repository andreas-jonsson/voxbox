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

package voxel

import "image/color"

type Image interface {
	Bounds() Box
	Set(x, y, z int, index uint8)
	Get(x, y, z int) uint8
}

func Blit(dst, src Image, dp Point, sr Box) {
	sr = sr.Intersect(src.Bounds())
	dr := Box{dp, sr.Size().Add(dp)}
	b := dst.Bounds().Intersect(dr)

	for z, sz := b.Min.Z, sr.Min.Z; z < b.Max.Z; z++ {
		for y, sy := b.Min.Y, sr.Min.Y; y < b.Max.Y; y++ {
			for x, sx := b.Min.X, sr.Min.X; x < b.Max.X; x++ {
				dst.Set(x, y, z, src.Get(sx, sy, sz))
				sx++
			}
			sy++
		}
		sz++
	}
}

type Op func(dst, src Image, dx, dy, dz, sx, sy, sz int)

func BlitOp(dst, src Image, dp Point, sr Box, op Op) {
	sr = sr.Intersect(src.Bounds())
	dr := Box{dp, sr.Size().Add(dp)}
	b := dst.Bounds().Intersect(dr)

	for z, sz := b.Min.Z, sr.Min.Z; z < b.Max.Z; z++ {
		for y, sy := b.Min.Y, sr.Min.Y; y < b.Max.Y; y++ {
			for x, sx := b.Min.X, sr.Min.X; x < b.Max.X; x++ {
				op(dst, src, x, y, z, sx, sy, sz)
				sx++
			}
			sy++
		}
		sz++
	}
}

type Paletted struct {
	bounds      Box
	Transformer func(x, y, z int) (int, int, int)
	Palette     color.Palette
	Data        []uint8
}

func noTransform(x, y, z int) (int, int, int) {
	return x, y, z
}

func NewPaletted(p color.Palette, b Box) *Paletted {
	img := &Paletted{Palette: p, Transformer: noTransform}
	img.SetBounds(b)
	return img
}

func (p *Paletted) Bounds() Box {
	return p.bounds
}

func (p *Paletted) SetBounds(b Box) {
	x, y, z := p.Transformer(b.Max.X, b.Max.Y, b.Max.Z)
	p.bounds = Box{ZP, Pt(x, y, z)}
	sz := b.Max.X * b.Max.Y * b.Max.Z
	p.Data = make([]uint8, sz)
}

func (p *Paletted) SetPalette(pal color.Palette) {
	p.Palette = pal
}

func (p *Paletted) Set(x, y, z int, index uint8) {
	x, y, z = p.Transformer(x, y, z)
	p.Data[p.Offset(x, y, z)] = index
}

func (p *Paletted) Get(x, y, z int) uint8 {
	return p.Data[p.Offset(x, y, z)]
}

func (p *Paletted) GetColor(x, y, z int) color.Color {
	return p.Palette[p.Get(x, y, z)]
}

func (p *Paletted) Offset(x, y, z int) int {
	return z*p.bounds.Max.X*p.bounds.Max.Y + y*p.bounds.Max.X + x
}

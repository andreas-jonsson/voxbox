// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package voxel

import (
	"fmt"
	"image/color"
)

type Point struct {
	X, Y, Z int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d,%d,%d)", p.X, p.Y, p.Z)
}

func (p Point) Add(q Point) Point {
	return Point{p.X + q.X, p.Y + q.Y, p.Z + q.Z}
}

func (p Point) Sub(q Point) Point {
	return Point{p.X - q.X, p.Y - q.Y, p.Z - q.Z}
}

func (p Point) Mul(k int) Point {
	return Point{p.X * k, p.Y * k, p.Z * k}
}

func (p Point) Div(k int) Point {
	return Point{p.X / k, p.Y / k, p.Z / k}
}

func (p Point) In(b Box) bool {
	return b.Min.X <= p.X && p.X < b.Max.X &&
		b.Min.Y <= p.Y && p.Y < b.Max.Y &&
		b.Min.Z <= p.Z && p.Z < b.Max.Z
}

func (p Point) Mod(b Box) Point {
	w, h, d := b.Dx(), b.Dy(), b.Dz()
	p = p.Sub(b.Min)
	p.X = p.X % w
	if p.X < 0 {
		p.X += w
	}
	p.Y = p.Y % h
	if p.Y < 0 {
		p.Y += h
	}
	p.Z = p.Z % d
	if p.Z < 0 {
		p.Z += d
	}
	return p.Add(b.Min)
}

func (p Point) Eq(q Point) bool {
	return p == q
}

var ZP Point

func Pt(X, Y, Z int) Point {
	return Point{X, Y, Z}
}

type Box struct {
	Min, Max Point
}

func (b Box) String() string {
	return b.Min.String() + "-" + b.Max.String()
}

func (b Box) Dx() int {
	return b.Max.X - b.Min.X
}

func (b Box) Dy() int {
	return b.Max.Y - b.Min.Y
}

func (b Box) Dz() int {
	return b.Max.Z - b.Min.Z
}

func (b Box) Size() Point {
	return Point{
		b.Max.X - b.Min.X,
		b.Max.Y - b.Min.Y,
		b.Max.Z - b.Min.Z,
	}
}

func (b Box) Add(p Point) Box {
	return Box{
		Point{b.Min.X + p.X, b.Min.Y + p.Y, b.Min.Z + p.Z},
		Point{b.Max.X + p.X, b.Max.Y + p.Y, b.Max.Z + p.Z},
	}
}

func (b Box) Sub(p Point) Box {
	return Box{
		Point{b.Min.X - p.X, b.Min.Y - p.Y, b.Min.Z - p.Z},
		Point{b.Max.X - p.X, b.Max.Y - p.Y, b.Max.Z - p.Z},
	}
}

func (b Box) Inset(n int) Box {
	if b.Dx() < 2*n {
		b.Min.X = (b.Min.X + b.Max.X) / 2
		b.Max.X = b.Min.X
	} else {
		b.Min.X += n
		b.Max.X -= n
	}
	if b.Dy() < 2*n {
		b.Min.Y = (b.Min.Y + b.Max.Y) / 2
		b.Max.Y = b.Min.Y
	} else {
		b.Min.Y += n
		b.Max.Y -= n
	}
	if b.Dz() < 2*n {
		b.Min.Z = (b.Min.Z + b.Max.Z) / 2
		b.Max.Z = b.Min.Z
	} else {
		b.Min.Z += n
		b.Max.Z -= n
	}
	return b
}

func (b Box) Intersect(s Box) Box {
	if b.Min.X < s.Min.X {
		b.Min.X = s.Min.X
	}
	if b.Min.Y < s.Min.Y {
		b.Min.Y = s.Min.Y
	}
	if b.Min.Z < s.Min.Z {
		b.Min.Z = s.Min.Z
	}
	if b.Max.X > s.Max.X {
		b.Max.X = s.Max.X
	}
	if b.Max.Y > s.Max.Y {
		b.Max.Y = s.Max.Y
	}
	if b.Max.Z > s.Max.Z {
		b.Max.Z = s.Max.Z
	}
	if b.Min.X > b.Max.X || b.Min.Y > b.Max.Y || b.Min.Z > b.Max.Z {
		return ZB
	}
	return b
}

func (b Box) Union(s Box) Box {
	if b.Empty() {
		return s
	}
	if s.Empty() {
		return b
	}
	if b.Min.X > s.Min.X {
		b.Min.X = s.Min.X
	}
	if b.Min.Y > s.Min.Y {
		b.Min.Y = s.Min.Y
	}
	if b.Min.Z > s.Min.Z {
		b.Min.Z = s.Min.Z
	}
	if b.Max.X < s.Max.X {
		b.Max.X = s.Max.X
	}
	if b.Max.Y < s.Max.Y {
		b.Max.Y = s.Max.Y
	}
	if b.Max.Z < s.Max.Z {
		b.Max.Z = s.Max.Z
	}
	return b
}

func (b Box) Empty() bool {
	return b.Min.X >= b.Max.X || b.Min.Y >= b.Max.Y || b.Min.Z >= b.Max.Z
}

func (b Box) Eq(s Box) bool {
	return b == s || b.Empty() && s.Empty()
}

func (b Box) Overlaps(s Box) bool {
	return !b.Empty() && !s.Empty() &&
		b.Min.X < s.Max.X && s.Min.X < b.Max.X &&
		b.Min.Y < s.Max.Y && s.Min.Y < b.Max.Y &&
		b.Min.Z < s.Max.Z && s.Min.Z < b.Max.Z
}

func (b Box) In(s Box) bool {
	if b.Empty() {
		return true
	}
	return s.Min.X <= b.Min.X && b.Max.X <= s.Max.X &&
		s.Min.Y <= b.Min.Y && b.Max.Y <= s.Max.Y &&
		s.Min.Z <= b.Min.Z && b.Max.Z <= s.Max.Z
}

func (b Box) Canon() Box {
	if b.Max.X < b.Min.X {
		b.Min.X, b.Max.X = b.Max.X, b.Min.X
	}
	if b.Max.Y < b.Min.Y {
		b.Min.Y, b.Max.Y = b.Max.Y, b.Min.Y
	}
	if b.Max.Z < b.Min.Z {
		b.Min.Z, b.Max.Z = b.Max.Z, b.Min.Z
	}
	return b
}

func (b Box) At(x, y, z int) color.Color {
	if (Point{x, y, z}).In(b) {
		return color.Opaque
	}
	return color.Transparent
}

func (b Box) Bounds() Box {
	return b
}

func (b Box) ColorModel() color.Model {
	return color.Alpha16Model
}

var ZB Box

func Bx(x0, y0, z0, x1, y1, z1 int) Box {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	if z0 > z1 {
		z0, z1 = z1, z0
	}
	return Box{Point{x0, y0, z0}, Point{x1, y1, z1}}
}

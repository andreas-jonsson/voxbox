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

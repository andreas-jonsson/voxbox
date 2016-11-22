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

package vox

import (
	"image"
	"image/color"
	"image/color/palette"
	"os"
	"testing"

	"github.com/andreas-jonsson/voxbox/voxel"
)

type voxelImage struct {
	data []*image.Paletted
}

func (img *voxelImage) SetBounds(b voxel.Box) {
	rect := image.Rect(0, 0, b.Max.X, b.Max.Y)
	img.data = make([]*image.Paletted, b.Max.Z)

	for i := 0; i < b.Max.Z; i++ {
		img.data[i] = image.NewPaletted(rect, palette.Plan9)
	}
}

func (img *voxelImage) SetPalette(pal color.Palette) {
	for _, layer := range img.data {
		layer.Palette = pal
	}
}

func (img *voxelImage) Set(x, y, z int, index uint8) {
	img.data[z].SetColorIndex(x, y, index)
}

func TestVox(t *testing.T) {
	if fp, err := os.Open("test.vox"); err == nil {
		defer fp.Close()

		var img voxelImage
		if err := Decode(fp, &img); err != nil {
			t.Error(err)
		}
	} else {
		t.Error(err)
	}
}
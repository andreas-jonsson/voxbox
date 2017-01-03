// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

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

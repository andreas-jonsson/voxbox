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
	"image/color/palette"
	"os"

	"github.com/andreas-jonsson/vox"
	"github.com/andreas-jonsson/voxbox/game"
)

type voxelImage struct {
	x, y, z int
	data    []uint8
	pal     color.Palette
}

func (img *voxelImage) SetSize(x, y, z int) {
	img.x, img.y, img.z = x, y, z

	img.data = make([]uint8, x*y*z)
	img.pal = palette.Plan9
}

func (img *voxelImage) SetPalette(pal color.Palette) {
	img.pal = pal
}

func (img *voxelImage) SetColorIndex(x, y, z int, index uint8) {
	img.data[img.offset(x, y, z)] = index
}

func (img *voxelImage) GetColorIndex(x, y, z int) uint8 {
	return img.data[img.offset(x, y, z)]
}

func (img *voxelImage) offset(x, y, z int) int {
	return z*img.x*img.y + y*img.x + x
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
	fp, err := os.Open("test.vox")
	if err != nil {
		return err
	}
	defer fp.Close()

	if err := vox.Read(fp, &s.img); err != nil {
		return err
	}
	return nil
}

func (s *playState) Exit(to game.GameState) error {
	return nil
}

func (s *playState) generateVoxel(c color.Color, x, y, z int) {

}

func (s *playState) Update(gctl game.GameControl) error {
	gctl.PollAll()

	img := &s.img
	pal := img.pal

	for z := 0; z < img.z; z++ {
		for y := 0; y < img.y; y++ {
			for x := 0; x < img.x; x++ {
				idx := img.GetColorIndex(x, y, z)
				if idx > 0 {
					s.generateVoxel(pal[idx], x, y, z)
				}
			}
		}
	}

	return nil
}

func (s *playState) Render() error {
	return nil
}

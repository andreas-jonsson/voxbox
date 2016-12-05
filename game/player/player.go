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

package player

import (
	"log"

	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/andreas-jonsson/warp/data"
	"github.com/ungerik/go3d/vec3"
)

type Player struct {
	forward    vec3.T
	renderFunc RenderFunc
	image      voxel.Paletted
}

type RenderFunc func(*Player, voxel.Image) error

func NewPlayer(rf RenderFunc) *Player {
	p := &Player{renderFunc: rf}

	fp, err := data.FS.Open("player.vox")
	if err != nil {
		log.Panicln(err)
	}
	defer fp.Close()

	p.image.Transformer = func(x, y, z int) (int, int, int) {
		return x, z, y
	}

	if err := vox.Decode(fp, &p.image); err != nil {
		log.Panicln(err)
	}

	return p
}

func (p *Player) Render() error {
	return p.renderFunc(p, &p.image)
}

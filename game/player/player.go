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

	"github.com/andreas-jonsson/voxbox/room"
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
	"github.com/andreas-jonsson/warp/data"
)

type Player struct {
	image voxel.Paletted
	view  voxel.Image
	room  *room.Room
	alive bool
}

func NewPlayer(view voxel.Image) *Player {
	p := &Player{view: view, alive: true}

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

func (p *Player) SetRoom(r *room.Room) {
	p.room = r
}

func (p *Player) blit(dst voxel.Image) {
	voxel.BlitOp(dst, &p.image, voxel.ZP, p.image.Bounds(), func(dst, src voxel.Image, dx, dy, dz, sx, sy, sz int) {
		c := src.Get(sx, sy, sz)
		if c > 0 {
			dst.Set(dx, dy, dz, c)
		}
	})
}

func (p *Player) Die() {
	if p.alive {
		p.alive = false

		// 	Do not wait for result.
		p.room.Send(func() {
			p.blit(p.room)
		})
	}
}

func (p *Player) Render() {
	if p.alive {
		p.blit(p.view)
	}
}

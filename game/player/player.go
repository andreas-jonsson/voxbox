// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package player

import (
	"log"

	"github.com/andreas-jonsson/voxbox/data"
	"github.com/andreas-jonsson/voxbox/room"
	"github.com/andreas-jonsson/voxbox/voxel"
	"github.com/andreas-jonsson/voxbox/voxel/vox"
)

type Player struct {
	image voxel.Paletted
	view  voxel.Image
	room  room.Interface
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

func (p *Player) SetRoom(r room.Interface) {
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
		p.room.Send(func(r *room.Room) {
			p.blit(r)
		})
	}
}

func (p *Player) Render() {
	if p.alive {
		p.blit(p.view)
	}
}

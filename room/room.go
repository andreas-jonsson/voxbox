// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package room

import (
	"errors"
	"image/color"
	"io"
	"time"

	"github.com/andreas-jonsson/voxbox/data"
	"github.com/andreas-jonsson/voxel/voxel"
	"github.com/andreas-jonsson/voxel/voxel/vox"
)

type Flag uint8

const (
	None     = 0x0
	Falling  = 0x80
	Attached = 0x40
)

const (
	invFalling  = 0x7F
	invAttached = 0xBF
)

const (
	attachedOrFalling     = Attached | Falling
	invAttachedAndFalling = invAttached & invFalling
)

const (
	sendBufferSize   = 128
	markTickDuration = 500 * time.Millisecond
)

/*
	var slide4Tab = [...]voxel.Point{
		voxel.Pt(1, -1, 0),
		voxel.Pt(-1, -1, 0),
		voxel.Pt(0, -1, 1),
		voxel.Pt(0, -1, -1),
	}
*/
var slideTab = [...]voxel.Point{
	voxel.Pt(-1, -1, -1),
	voxel.Pt(0, -1, -1),
	voxel.Pt(1, -1, -1),
	voxel.Pt(-1, -1, 0),
	voxel.Pt(1, -1, 0),
	voxel.Pt(-1, -1, 1),
	voxel.Pt(0, -1, 1),
	voxel.Pt(1, -1, 1),
}

/*
var slideTab = [...]voxel.Point{
	voxel.Pt(-1, -1, -1),
	voxel.Pt(0, -1, -1),
	voxel.Pt(1, -1, -1),
	voxel.Pt(-1, -1, 0),
	voxel.Pt(1, -1, 0),
	voxel.Pt(-1, -1, 1),
	voxel.Pt(0, -1, 1),
	voxel.Pt(1, -1, 1),

	voxel.Pt(-1, 0, -1),
	voxel.Pt(0, 0, -1),
	voxel.Pt(1, 0, -1),
	voxel.Pt(-1, 0, 0),
	voxel.Pt(1, 0, 0),
	voxel.Pt(-1, 0, 1),
	voxel.Pt(0, 0, 1),
	voxel.Pt(1, 0, 1),
}
*/

const slideTabLen = uint32(len(slideTab))

type Room struct {
	loadPos, size voxel.Point
	bounds        voxel.Box
	flipYZ        bool
	flags         Flag
	data          []uint8

	stepTicker, markTicker *time.Ticker

	funcChan chan func()
	stopChan chan struct{}
}

type Interface interface {
	Send(f func(*Room)) <-chan struct{}
	Clear()
	Bounds() voxel.Box
	BlitToView(dst voxel.ImageData, dp voxel.Point, sr voxel.Box) <-chan struct{}
	Destroy()
}

func NewRoom(size voxel.Point, simSpeed time.Duration) *Room {
	return &Room{
		stopChan:   make(chan struct{}),
		funcChan:   make(chan func(), sendBufferSize),
		stepTicker: time.NewTicker(simSpeed),
		markTicker: time.NewTicker(markTickDuration),
		size:       size,
		bounds:     voxel.Box{Min: voxel.ZP, Max: size},
		data:       make([]uint8, size.X*size.Y*size.Z),
	}
}

func (r *Room) Send(f func(*Room)) <-chan struct{} {
	cbChan := make(chan struct{}, 1)
	r.funcChan <- func() {
		f(r)
		cbChan <- struct{}{}
	}
	return cbChan
}

func (r *Room) Destroy() {
	r.stopChan <- struct{}{}
	r.stepTicker.Stop()
	r.markTicker.Stop()

}

func (r *Room) Clear() {
	r.Send(func(r *Room) {
		for i := range r.data {
			r.data[i] = 0
		}
	})
}

func (r *Room) Start() Interface {
	go func(r *Room) {
		for {
			select {
			case <-r.markTicker.C:
				r.markPhase()
			case <-r.stepTicker.C:
				r.stepPhase()
			case f := <-r.funcChan:
				f()
			case <-r.stopChan:
				break
			}
		}
	}(r)

	return r
}

func (r *Room) markPhase() {
	box := r.bounds
	normals := [...]voxel.Point{
		voxel.Pt(1, 0, 0),
		voxel.Pt(-1, 0, 0),
		voxel.Pt(0, -1, 0),
		voxel.Pt(0, 1, 0),
		voxel.Pt(0, 0, 1),
		voxel.Pt(0, 0, -1),
	}

	for z := 0; z < r.size.Z; z++ {
		for y := 0; y < r.size.Y; y++ {
			for x := 0; x < r.size.X; x++ {
				vIdx := r.offset(x, y, z)
				v := r.data[vIdx]

				if v == 0 {
					continue
				}

				if y == 0 {
					r.data[vIdx] = (v & invFalling) | Attached
					continue
				}

				if v&Falling != 0 {
					continue
				}

				v &= invAttached
				p := voxel.Point{X: x, Y: y, Z: z}

				for i := 0; i < 6; i++ {
					np := p.Add(normals[i])
					if np.In(box) {
						nIdx := r.offset(np.X, np.Y, np.Z)
						if r.data[nIdx]&Attached != 0 {
							v |= Attached
							break
						}
					}
				}

				r.data[vIdx] = v
			}
		}
	}
}

func (r *Room) stepPhase() {
	var randSeed uint32 // Could use stdlib rand but this is faster.
	box := r.bounds

	for y := 1; y < r.size.Y; y++ {
		for z := 0; z < r.size.Z; z++ {
			for x := 0; x < r.size.X; x++ {
				vIdx := r.offset(x, y, z)
				v := r.data[vIdx]

				if v == 0 || v&Attached != 0 {
					continue
				}

				nIdx := r.offset(x, y-1, z)
				nv := r.data[nIdx]

				if nv == 0 {
					r.data[vIdx] = 0
					r.data[nIdx] = v | Falling
				} else {
					randSeed = randSeed*1664525 + 1013904223
					rnd := randSeed >> 24

					sn := slideTab[rnd%slideTabLen]
					snp := voxel.Point{X: x + sn.X, Y: y + sn.Y, Z: z + sn.Z}

					if !snp.In(box) {
						r.data[vIdx] = (v & invAttachedAndFalling) | (nv & attachedOrFalling)
						continue
					}

					snIdx := r.offset(snp.X, snp.Y, snp.Z)
					snv := r.data[snIdx]

					if snv == 0 {
						r.data[vIdx] = 0
						r.data[snIdx] = v | Falling
					} else {
						r.data[vIdx] = (v & invAttachedAndFalling) | (nv & attachedOrFalling)
					}
				}
			}
		}
	}
}

func (r *Room) BlitToView(dst voxel.ImageData, dp voxel.Point, sr voxel.Box) <-chan struct{} {
	return r.Send(func(r *Room) {
		sr = sr.Intersect(r.Bounds())
		dr := voxel.Box{Min: dp, Max: sr.Size().Add(dp)}
		b := dst.Bounds().Intersect(dr)

		blockSize := b.Max.X - b.Min.X
		dstSize := dst.Bounds().Max
		srcSize := r.Bounds().Max
		dstData := dst.Data()

		for z, sz := b.Min.Z, sr.Min.Z; z < b.Max.Z; z++ {
			for y, sy := b.Min.Y, sr.Min.Y; y < b.Max.Y; y++ {

				dstStart := z*dstSize.X*dstSize.Y + y*dstSize.X + b.Min.X
				dstSlice := dstData[dstStart : dstStart+blockSize]
				srcStart := sz*srcSize.X*srcSize.Y + sy*srcSize.X + sr.Min.X

				for i, v := range r.data[srcStart : srcStart+blockSize] {
					dstSlice[i] = v & invAttachedAndFalling
				}

				sy++
			}
			sz++
		}
	})
}

func (r *Room) Bounds() voxel.Box {
	return r.bounds
}

func (r *Room) SetBounds(b voxel.Box) {
}

func (r *Room) SetPalette(pal color.Palette) {
}

func (r *Room) Set(x, y, z int, index uint8) {
	var cIdx uint8
	if index > 0 {
		cIdx = (invAttachedAndFalling & index) | uint8(r.flags)
	}

	if r.flipYZ {
		x += r.loadPos.X
		z += r.loadPos.Y
		y += r.loadPos.Z

		if x < r.size.X && z < r.size.Y && y < r.size.Z {
			r.data[r.offset(x, z, y)] = cIdx
		}
	} else {
		x += r.loadPos.X
		y += r.loadPos.Y
		z += r.loadPos.Z

		if x < r.size.X && y < r.size.Y && z < r.size.Z {
			r.data[r.offset(x, y, z)] = cIdx
		}
	}
}

func (r *Room) Get(x, y, z int) uint8 {
	return r.data[r.offset(x, y, z)] & invAttachedAndFalling
}

func (r *Room) offset(x, y, z int) int {
	return z*r.size.X*r.size.Y + y*r.size.X + x
}

func (r *Room) LoadVOXFile(file string, at voxel.Point, flag Flag) error {
	fp, err := data.FS.Open(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	return r.LoadVOX(fp, at, flag)
}

func (r *Room) LoadVOX(reader io.Reader, at voxel.Point, flag Flag) error {
	if flag&Attached != 0 && flag&Falling != 0 {
		return errors.New("Attached and Falling flag can't be set at the same time")
	}

	lp := r.loadPos
	flip := r.flipYZ
	fl := r.flags

	r.loadPos = at
	r.flipYZ = true
	r.flags = flag

	err := vox.Decode(reader, r)

	r.flags = fl
	r.loadPos = lp
	r.flipYZ = flip
	return err
}

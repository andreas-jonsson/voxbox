// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package vox

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"

	"github.com/andreas-jonsson/voxbox/voxel"
)

const (
	voxMagic   = "VOX "
	voxVersion = 150
)

const (
	mainChunkID    = "MAIN"
	sizeShunkID    = "SIZE"
	voxelChunkID   = "XYZI"
	paletteChunkID = "RGBA"
)

var (
	ErrInvalidFile      = Error{"invalid file", nil}
	ErrInvalidVersion   = Error{"invalid version", nil}
	ErrInvalidChunk     = Error{"invalid chunk", nil}
	ErrInvalidMainChunk = Error{"invalid main chunk", nil}
)

type Error struct {
	err   string
	inner error
}

func (e Error) Error() string {
	if e.inner == nil {
		return e.err
	}
	return fmt.Sprintf("%s: %v", e.err, e.inner)
}

func (e Error) with(inner error) Error {
	return Error{e.err, inner}
}

type Image interface {
	SetBounds(b voxel.Box)
	SetPalette(pal color.Palette)
	Set(x, y, z int, index uint8)
}

type (
	voxHeader struct {
		Magic   [4]byte
		Version [4]byte
	}

	chunkHeader struct {
		Id [4]byte
		DataSize,
		ChildrenSize uint32
	}
)

func Decode(reader io.Reader, img Image) error {
	var fileHeader voxHeader
	if err := binary.Read(reader, binary.LittleEndian, &fileHeader); err != nil {
		return ErrInvalidFile.with(err)
	}

	if string(fileHeader.Magic[:]) != voxMagic {
		return ErrInvalidFile
	}

	if fileHeader.Version[0] != voxVersion {
		return ErrInvalidVersion
	}

	var header chunkHeader
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return ErrInvalidMainChunk.with(err)
	}

	if string(header.Id[:]) != mainChunkID {
		return ErrInvalidMainChunk
	}

	var (
		hasPalette bool
		numBytes   uint32
	)

	childrenSize := header.ChildrenSize
	for numBytes < childrenSize {
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			return ErrInvalidFile.with(err)
		}
		numBytes += 12

		switch string(header.Id[:]) {
		case sizeShunkID:
			var size [3]uint32
			if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
				return ErrInvalidChunk.with(err)
			}

			numBytes += 12
			img.SetBounds(voxel.Bx(0, 0, 0, int(size[0]), int(size[1]), int(size[2])))
		case paletteChunkID:
			palette := make(color.Palette, 256)
			for i := range palette {
				var c color.RGBA
				if err := binary.Read(reader, binary.LittleEndian, &c); err != nil {
					return ErrInvalidChunk.with(err)
				}
				palette[i] = c
			}

			hasPalette = true
			numBytes += 16 * 256
			img.SetPalette(palette)
		case voxelChunkID:
			var numVoxels uint32
			if err := binary.Read(reader, binary.LittleEndian, &numVoxels); err != nil {
				return ErrInvalidChunk.with(err)
			}
			numBytes += 4

			for i := uint32(0); i < numVoxels; i++ {
				var voxel [4]byte
				if err := binary.Read(reader, binary.LittleEndian, &voxel); err != nil {
					return ErrInvalidChunk.with(err)
				}
				img.Set(int(voxel[0]), int(voxel[1]), int(voxel[2]), voxel[3])
			}
			numBytes += 4 * numVoxels
		default:
			sz := header.DataSize + header.ChildrenSize
			if _, err := reader.Read(make([]byte, sz)); err != nil {
				return ErrInvalidFile.with(err)
			}
			numBytes += sz
		}
	}

	if !hasPalette {
		img.SetPalette(defaultPalette[:])
	}

	return nil
}

var defaultPalette = [256]color.Color{
	color.RGBA{255, 255, 255, 255},
	color.RGBA{255, 255, 204, 255},
	color.RGBA{255, 255, 153, 255},
	color.RGBA{255, 255, 102, 255},
	color.RGBA{255, 255, 51, 255},
	color.RGBA{255, 255, 0, 255},
	color.RGBA{255, 204, 255, 255},
	color.RGBA{255, 204, 204, 255},
	color.RGBA{255, 204, 153, 255},
	color.RGBA{255, 204, 102, 255},
	color.RGBA{255, 204, 51, 255},
	color.RGBA{255, 204, 0, 255},
	color.RGBA{255, 153, 255, 255},
	color.RGBA{255, 153, 204, 255},
	color.RGBA{255, 153, 153, 255},
	color.RGBA{255, 153, 102, 255},
	color.RGBA{255, 153, 51, 255},
	color.RGBA{255, 153, 0, 255},
	color.RGBA{255, 102, 255, 255},
	color.RGBA{255, 102, 204, 255},
	color.RGBA{255, 102, 153, 255},
	color.RGBA{255, 102, 102, 255},
	color.RGBA{255, 102, 51, 255},
	color.RGBA{255, 102, 0, 255},
	color.RGBA{255, 51, 255, 255},
	color.RGBA{255, 51, 204, 255},
	color.RGBA{255, 51, 153, 255},
	color.RGBA{255, 51, 102, 255},
	color.RGBA{255, 51, 51, 255},
	color.RGBA{255, 51, 0, 255},
	color.RGBA{255, 0, 255, 255},
	color.RGBA{255, 0, 204, 255},
	color.RGBA{255, 0, 153, 255},
	color.RGBA{255, 0, 102, 255},
	color.RGBA{255, 0, 51, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{204, 255, 255, 255},
	color.RGBA{204, 255, 204, 255},
	color.RGBA{204, 255, 153, 255},
	color.RGBA{204, 255, 102, 255},
	color.RGBA{204, 255, 51, 255},
	color.RGBA{204, 255, 0, 255},
	color.RGBA{204, 204, 255, 255},
	color.RGBA{204, 204, 204, 255},
	color.RGBA{204, 204, 153, 255},
	color.RGBA{204, 204, 102, 255},
	color.RGBA{204, 204, 51, 255},
	color.RGBA{204, 204, 0, 255},
	color.RGBA{204, 153, 255, 255},
	color.RGBA{204, 153, 204, 255},
	color.RGBA{204, 153, 153, 255},
	color.RGBA{204, 153, 102, 255},
	color.RGBA{204, 153, 51, 255},
	color.RGBA{204, 153, 0, 255},
	color.RGBA{204, 102, 255, 255},
	color.RGBA{204, 102, 204, 255},
	color.RGBA{204, 102, 153, 255},
	color.RGBA{204, 102, 102, 255},
	color.RGBA{204, 102, 51, 255},
	color.RGBA{204, 102, 0, 255},
	color.RGBA{204, 51, 255, 255},
	color.RGBA{204, 51, 204, 255},
	color.RGBA{204, 51, 153, 255},
	color.RGBA{204, 51, 102, 255},
	color.RGBA{204, 51, 51, 255},
	color.RGBA{204, 51, 0, 255},
	color.RGBA{204, 0, 255, 255},
	color.RGBA{204, 0, 204, 255},
	color.RGBA{204, 0, 153, 255},
	color.RGBA{204, 0, 102, 255},
	color.RGBA{204, 0, 51, 255},
	color.RGBA{204, 0, 0, 255},
	color.RGBA{153, 255, 255, 255},
	color.RGBA{153, 255, 204, 255},
	color.RGBA{153, 255, 153, 255},
	color.RGBA{153, 255, 102, 255},
	color.RGBA{153, 255, 51, 255},
	color.RGBA{153, 255, 0, 255},
	color.RGBA{153, 204, 255, 255},
	color.RGBA{153, 204, 204, 255},
	color.RGBA{153, 204, 153, 255},
	color.RGBA{153, 204, 102, 255},
	color.RGBA{153, 204, 51, 255},
	color.RGBA{153, 204, 0, 255},
	color.RGBA{153, 153, 255, 255},
	color.RGBA{153, 153, 204, 255},
	color.RGBA{153, 153, 153, 255},
	color.RGBA{153, 153, 102, 255},
	color.RGBA{153, 153, 51, 255},
	color.RGBA{153, 153, 0, 255},
	color.RGBA{153, 102, 255, 255},
	color.RGBA{153, 102, 204, 255},
	color.RGBA{153, 102, 153, 255},
	color.RGBA{153, 102, 102, 255},
	color.RGBA{153, 102, 51, 255},
	color.RGBA{153, 102, 0, 255},
	color.RGBA{153, 51, 255, 255},
	color.RGBA{153, 51, 204, 255},
	color.RGBA{153, 51, 153, 255},
	color.RGBA{153, 51, 102, 255},
	color.RGBA{153, 51, 51, 255},
	color.RGBA{153, 51, 0, 255},
	color.RGBA{153, 0, 255, 255},
	color.RGBA{153, 0, 204, 255},
	color.RGBA{153, 0, 153, 255},
	color.RGBA{153, 0, 102, 255},
	color.RGBA{153, 0, 51, 255},
	color.RGBA{153, 0, 0, 255},
	color.RGBA{102, 255, 255, 255},
	color.RGBA{102, 255, 204, 255},
	color.RGBA{102, 255, 153, 255},
	color.RGBA{102, 255, 102, 255},
	color.RGBA{102, 255, 51, 255},
	color.RGBA{102, 255, 0, 255},
	color.RGBA{102, 204, 255, 255},
	color.RGBA{102, 204, 204, 255},
	color.RGBA{102, 204, 153, 255},
	color.RGBA{102, 204, 102, 255},
	color.RGBA{102, 204, 51, 255},
	color.RGBA{102, 204, 0, 255},
	color.RGBA{102, 153, 255, 255},
	color.RGBA{102, 153, 204, 255},
	color.RGBA{102, 153, 153, 255},
	color.RGBA{102, 153, 102, 255},
	color.RGBA{102, 153, 51, 255},
	color.RGBA{102, 153, 0, 255},
	color.RGBA{102, 102, 255, 255},
	color.RGBA{102, 102, 204, 255},
	color.RGBA{102, 102, 153, 255},
	color.RGBA{102, 102, 102, 255},
	color.RGBA{102, 102, 51, 255},
	color.RGBA{102, 102, 0, 255},
	color.RGBA{102, 51, 255, 255},
	color.RGBA{102, 51, 204, 255},
	color.RGBA{102, 51, 153, 255},
	color.RGBA{102, 51, 102, 255},
	color.RGBA{102, 51, 51, 255},
	color.RGBA{102, 51, 0, 255},
	color.RGBA{102, 0, 255, 255},
	color.RGBA{102, 0, 204, 255},
	color.RGBA{102, 0, 153, 255},
	color.RGBA{102, 0, 102, 255},
	color.RGBA{102, 0, 51, 255},
	color.RGBA{102, 0, 0, 255},
	color.RGBA{51, 255, 255, 255},
	color.RGBA{51, 255, 204, 255},
	color.RGBA{51, 255, 153, 255},
	color.RGBA{51, 255, 102, 255},
	color.RGBA{51, 255, 51, 255},
	color.RGBA{51, 255, 0, 255},
	color.RGBA{51, 204, 255, 255},
	color.RGBA{51, 204, 204, 255},
	color.RGBA{51, 204, 153, 255},
	color.RGBA{51, 204, 102, 255},
	color.RGBA{51, 204, 51, 255},
	color.RGBA{51, 204, 0, 255},
	color.RGBA{51, 153, 255, 255},
	color.RGBA{51, 153, 204, 255},
	color.RGBA{51, 153, 153, 255},
	color.RGBA{51, 153, 102, 255},
	color.RGBA{51, 153, 51, 255},
	color.RGBA{51, 153, 0, 255},
	color.RGBA{51, 102, 255, 255},
	color.RGBA{51, 102, 204, 255},
	color.RGBA{51, 102, 153, 255},
	color.RGBA{51, 102, 102, 255},
	color.RGBA{51, 102, 51, 255},
	color.RGBA{51, 102, 0, 255},
	color.RGBA{51, 51, 255, 255},
	color.RGBA{51, 51, 204, 255},
	color.RGBA{51, 51, 153, 255},
	color.RGBA{51, 51, 102, 255},
	color.RGBA{51, 51, 51, 255},
	color.RGBA{51, 51, 0, 255},
	color.RGBA{51, 0, 255, 255},
	color.RGBA{51, 0, 204, 255},
	color.RGBA{51, 0, 153, 255},
	color.RGBA{51, 0, 102, 255},
	color.RGBA{51, 0, 51, 255},
	color.RGBA{51, 0, 0, 255},
	color.RGBA{0, 255, 255, 255},
	color.RGBA{0, 255, 204, 255},
	color.RGBA{0, 255, 153, 255},
	color.RGBA{0, 255, 102, 255},
	color.RGBA{0, 255, 51, 255},
	color.RGBA{0, 255, 0, 255},
	color.RGBA{0, 204, 255, 255},
	color.RGBA{0, 204, 204, 255},
	color.RGBA{0, 204, 153, 255},
	color.RGBA{0, 204, 102, 255},
	color.RGBA{0, 204, 51, 255},
	color.RGBA{0, 204, 0, 255},
	color.RGBA{0, 153, 255, 255},
	color.RGBA{0, 153, 204, 255},
	color.RGBA{0, 153, 153, 255},
	color.RGBA{0, 153, 102, 255},
	color.RGBA{0, 153, 51, 255},
	color.RGBA{0, 153, 0, 255},
	color.RGBA{0, 102, 255, 255},
	color.RGBA{0, 102, 204, 255},
	color.RGBA{0, 102, 153, 255},
	color.RGBA{0, 102, 102, 255},
	color.RGBA{0, 102, 51, 255},
	color.RGBA{0, 102, 0, 255},
	color.RGBA{0, 51, 255, 255},
	color.RGBA{0, 51, 204, 255},
	color.RGBA{0, 51, 153, 255},
	color.RGBA{0, 51, 102, 255},
	color.RGBA{0, 51, 51, 255},
	color.RGBA{0, 51, 0, 255},
	color.RGBA{0, 0, 255, 255},
	color.RGBA{0, 0, 204, 255},
	color.RGBA{0, 0, 153, 255},
	color.RGBA{0, 0, 102, 255},
	color.RGBA{0, 0, 51, 255},
	color.RGBA{238, 0, 0, 255},
	color.RGBA{221, 0, 0, 255},
	color.RGBA{187, 0, 0, 255},
	color.RGBA{170, 0, 0, 255},
	color.RGBA{136, 0, 0, 255},
	color.RGBA{119, 0, 0, 255},
	color.RGBA{85, 0, 0, 255},
	color.RGBA{68, 0, 0, 255},
	color.RGBA{34, 0, 0, 255},
	color.RGBA{17, 0, 0, 255},
	color.RGBA{0, 238, 0, 255},
	color.RGBA{0, 221, 0, 255},
	color.RGBA{0, 187, 0, 255},
	color.RGBA{0, 170, 0, 255},
	color.RGBA{0, 136, 0, 255},
	color.RGBA{0, 119, 0, 255},
	color.RGBA{0, 85, 0, 255},
	color.RGBA{0, 68, 0, 255},
	color.RGBA{0, 34, 0, 255},
	color.RGBA{0, 17, 0, 255},
	color.RGBA{0, 0, 238, 255},
	color.RGBA{0, 0, 221, 255},
	color.RGBA{0, 0, 187, 255},
	color.RGBA{0, 0, 170, 255},
	color.RGBA{0, 0, 136, 255},
	color.RGBA{0, 0, 119, 255},
	color.RGBA{0, 0, 85, 255},
	color.RGBA{0, 0, 68, 255},
	color.RGBA{0, 0, 34, 255},
	color.RGBA{0, 0, 17, 255},
	color.RGBA{238, 238, 238, 255},
	color.RGBA{221, 221, 221, 255},
	color.RGBA{187, 187, 187, 255},
	color.RGBA{170, 170, 170, 255},
	color.RGBA{136, 136, 136, 255},
	color.RGBA{119, 119, 119, 255},
	color.RGBA{85, 85, 85, 255},
	color.RGBA{68, 68, 68, 255},
	color.RGBA{34, 34, 34, 255},
	color.RGBA{17, 17, 17, 255},
	color.RGBA{0, 0, 0, 255},
}

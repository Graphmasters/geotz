/*
Copyright 2014 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package tile holds the parts of the table format that the lookup code and
// the table generator both have to agree on: how a tile is identified, and
// which leaf index means "no timezone here".
//
// It exists so that the format has exactly one definition. The generator lives
// in its own Go module (see the gen directory) so that its heavyweight
// dependencies stay out of this one, which puts it beyond the reach of
// unexported identifiers — but not beyond an internal package.
package tile

// A Key is a packed 32 bit tile identifier:
//
//	bits 31-29: tile size; the tile is 8<<n pixels square, for n = 0 to 5
//	bit 28:     unused
//	bits 27-14: y position, in tiles
//	bits 13-0:  x position, in tiles
//
// Keys sort by size then position, which is what lets a lookup binary-search a
// zoom level for the tile containing a pixel.
type Key uint32

// NewKey packs a tile size exponent (0 to 5, for tiles of 8 to 256 pixels
// square) and a tile position into a Key.
func NewKey(size uint8, x, y uint16) Key {
	return Key(size&7)<<28 |
		Key(y&(1<<14-1))<<14 |
		Key(x&(1<<14-1))
}

// Size reports the tile's size exponent: the tile is 8<<Size pixels square.
func (v Key) Size() uint8 {
	return byte(v >> 28)
}

// X reports the tile's x position, in tiles.
func (v Key) X() uint16 {
	return uint16(v & (1<<14 - 1))
}

// Y reports the tile's y position, in tiles.
func (v Key) Y() uint16 {
	return uint16((v >> 14) & (1<<14 - 1))
}

// OceanIndex is the leaf index that means the pixel belongs to no timezone —
// open water in a dataset that does not cover the sea, or a gap between
// polygons. It is deliberately the largest uint16, so it can never collide
// with a real leaf index.
const OceanIndex uint16 = 0xffff

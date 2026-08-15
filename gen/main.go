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

// Command gen rasterizes a timezone-boundary-builder shapefile and writes the
// lookup tables that github.com/Graphmasters/geotz compiles in.
//
// It lives in its own Go module so that its dependencies — a shapefile reader
// and a rasterizer — stay out of the module people actually import, which has
// none. See ../README.md; the usual way to run it is `make z_gen_tables.go`
// from the repository root.
package main

import (
	"bufio"
	"flag"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/golang/freetype/raster"
	shp "github.com/jonas-p/go-shp"
	"golang.org/x/image/math/fixed"
)

var (
	flagWriteImage = flag.Bool("write_image", false, "Write out regions.png, a debug rendering of the rasterized world.")
	flagScale      = flag.Float64("scale", 32, "Scaling factor. This many pixels wide & tall per degree (e.g. scale 1 is 360 x 180). Increasingly this code assumes a scale of 32, though.")
	flagShapefile  = flag.String("shapefile", "../world/combined-shapefile-with-oceans-now.shp", "Path to the timezone-boundary-builder shapefile to rasterize. See the Makefile.")
	flagZoneField  = flag.String("zone_field", "tzid", "Name of the shapefile attribute holding the IANA timezone name.")
	flagTZRelease  = flag.String("tz_release", "2026c", "timezone-boundary-builder release the shapefile comes from. Recorded in the generated file for attribution.")
	flagOutput     = flag.String("o", "../z_gen_tables.go", "Where to write the generated tables.")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	im, zoneOfColor := worldImage()
	src := buildTables(im, zoneOfColor)
	if err := os.WriteFile(*flagOutput, src, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s (%d bytes)", *flagOutput, len(src))
}

const alphaErased = 22 // magic alpha value to mean tile's been erased

// worldImage rasterizes the timezone-boundary-builder shapefile into an
// equirectangular image in which a pixel's color identifies its timezone.
//
// The returned zoneOfColor always has A == 256.
func worldImage() (im *image.RGBA, zoneOfColor map[color.RGBA]string) {
	scale := *flagScale
	width := int(scale * 360)
	height := int(scale * 180)

	im = image.NewRGBA(image.Rect(0, 0, width, height))
	zoneOfColor = map[color.RGBA]string{}
	tab := crc32.MakeTable(crc32.IEEE + 1)

	painter := raster.NewRGBAPainter(im)
	mono := raster.NewMonochromePainter(painter)
	r := raster.NewRasterizer(width, height)
	// One ESRI polygon record holds many rings: outer rings and, wound the
	// other way, the holes in them. Adding every ring to a single path and
	// filling it in one pass is what cuts those holes back out — which this
	// dataset needs badly, since each ocean zone is one polygon with a hole
	// punched in it for every island it wraps around.
	//
	// The fill rule stays at the default even-odd rather than non-zero
	// winding, because even-odd takes the absolute coverage per pixel and so
	// does not care which way a ring is wound. That matters: the spec's
	// clockwise-is-outer convention is stated in map coordinates, and the
	// projection below flips Y. Both rules happen to produce byte-identical
	// tables from the 2026c data, so this is insurance against a future
	// release with sloppier winding, not a correction.
	r.UseNonZeroWinding = false

	// point converts a shapefile coordinate in degrees to the rasterizer's
	// 26.6 fixed point pixel space, sub-pixel part included: the monochrome
	// painter fills a pixel once it is at least half covered, so the extra
	// precision is what puts a border on the correct side of a pixel.
	point := func(p shp.Point) fixed.Point26_6 {
		return fixed.Point26_6{
			X: fixed.Int26_6((p.X + 180) * scale * 64),
			Y: fixed.Int26_6((90 - p.Y) * scale * 64),
		}
	}

	drawPolygon := func(col color.RGBA, p *shp.Polygon) {
		r.Clear()
		painter.SetColor(col)
		for i, start := range p.Parts {
			end := int32(len(p.Points))
			if i+1 < len(p.Parts) {
				end = p.Parts[i+1]
			}
			ring := p.Points[start:end]
			if len(ring) < 3 {
				continue
			}
			r.Start(point(ring[0]))
			for _, pt := range ring[1:] {
				r.Add1(point(pt))
			}
			r.Add1(point(ring[0])) // close the ring
		}
		r.Rasterize(mono)
	}

	sr, err := shp.Open(*flagShapefile)
	if err != nil {
		log.Fatalf("Error opening %s: %v; run `make world` to fetch it from "+
			"https://github.com/evansiroky/timezone-boundary-builder/releases", *flagShapefile, err)
	}
	defer func() {
		if err := sr.Close(); err != nil {
			log.Printf("closing %s: %v", *flagShapefile, err)
		}
	}()

	zoneField := -1
	var fieldNames []string
	for i, f := range sr.Fields() {
		fieldNames = append(fieldNames, f.String())
		if f.String() == *flagZoneField {
			zoneField = i
		}
	}
	if zoneField == -1 {
		log.Fatalf("%s has no %q attribute; it has %q", *flagShapefile, *flagZoneField, fieldNames)
	}

	records := 0
	for sr.Next() {
		i, s := sr.Shape()
		p, ok := s.(*shp.Polygon)
		if !ok {
			log.Fatalf("Unknown shape %T", s)
		}
		zoneName := sr.ReadAttribute(i, zoneField)
		if _, err := time.LoadLocation(zoneName); err != nil {
			log.Fatalf("Failed to load: %v (%v)", zoneName, err)
		}
		hash := crc32.Checksum([]byte(zoneName), tab)
		col := color.RGBA{uint8(hash >> 24), uint8(hash >> 16), uint8(hash >> 8), 255}
		if name, ok := zoneOfColor[col]; ok {
			if name != zoneName {
				log.Fatalf("Color %+v dup: %s and %s", col, name, zoneName)
			}
		} else {
			zoneOfColor[col] = zoneName
		}
		records++
		drawPolygon(col, p)
	}
	if err := sr.Err(); err != nil {
		log.Fatalf("Error reading %s: %v", *flagShapefile, err)
	}
	log.Printf("Rasterized %d records covering %d zones from %s", records, len(zoneOfColor), *flagShapefile)

	// This dataset covers the oceans too, so every pixel should have been
	// painted. Anything left transparent is a gap in the data (or a bug), and
	// will look up as the empty string.
	uncovered := 0
	for i := 3; i < len(im.Pix); i += 4 {
		if im.Pix[i] == 0 {
			uncovered++
		}
	}
	log.Printf("%d of %d pixels (%.5f%%) are covered by no zone",
		uncovered, width*height, 100*float64(uncovered)/float64(width*height))
	return
}

func saveToPNGFile(filePath string, m image.Image) {
	log.Printf("Encoding image %s ...", filePath)
	f, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
	}
	b := bufio.NewWriter(f)
	if err := png.Encode(b, m); err != nil {
		_ = f.Close() // about to exit; the write error is the interesting one
		log.Fatal(err)
	}
	if err := b.Flush(); err != nil {
		_ = f.Close() // about to exit; the write error is the interesting one
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s", filePath)
}

func cloneImage(i *image.RGBA) *image.RGBA {
	i2 := new(image.RGBA)
	*i2 = *i
	i2.Pix = make([]uint8, len(i.Pix))
	copy(i2.Pix, i.Pix)
	return i2
}

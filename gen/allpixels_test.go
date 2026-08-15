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

package main

import (
	"image/color"
	"os"
	"testing"

	"github.com/Graphmasters/geotz"
)

// TestAllPixels rasterizes the shapefile again and asks the published package
// about every single pixel of it. It is the check that the compression from
// image to tables lost nothing: 66 million lookups, each of which must return
// exactly the zone the polygon painted there.
//
// It needs the shapefile, so it skips unless `make world` has been run.
func TestAllPixels(t *testing.T) {
	if _, err := os.Stat(*flagShapefile); err != nil {
		t.Skipf("%s not present; run `make world` from the repository root first", *flagShapefile)
	}

	im, zoneOfColor := worldImage()
	w := im.Bounds().Max.X
	h := im.Bounds().Max.Y

	// Pixel (x, y) covers the coordinates that floor to it; aim at its centre.
	scale := *flagScale
	half := 0.5 / scale

	total, fail := 0, 0
	for y := 0; y < h; y++ {
		lat := 90 - float64(y)/scale - half
		for x := 0; x < w; x++ {
			pix := im.Pix[im.PixOffset(x, y):]
			if pix[3] == 0 {
				continue // no polygon covered this pixel
			}
			total++
			want := zoneOfColor[color.RGBA{R: pix[0], G: pix[1], B: pix[2], A: 255}]
			long := float64(x)/scale - 180 + half
			if got := geotz.LookupZoneName(lat, long); got != want {
				fail++
				if fail <= 10 {
					t.Errorf("LookupZoneName(%v, %v) [pixel %d,%d] = %q; want %q", lat, long, x, y, got, want)
				}
			}
		}
	}
	if fail > 10 {
		t.Errorf("%d failures in total", fail)
	}
	t.Logf("%d pixels tested; %d failures", total, fail)
}

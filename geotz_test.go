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

package geotz

import (
	"testing"
	"time"
)

// TestLookupTellsTheTime checks the property the package actually promises:
// the name it returns keeps the same time as the zone the place really is in.
// It deliberately does not pin the names themselves — the tables are built from
// the "now" dataset, where each region is labelled with one representative of
// all zones that currently keep identical time, and which member gets to be the
// representative is upstream's choice and changes between releases.
func TestLookupTellsTheTime(t *testing.T) {
	cases := []struct {
		place     string
		lat, long float64
		zone      string // the zone the place is really in
	}{
		{"Hannover", 52.3759, 9.7320, "Europe/Berlin"},
		{"London", 51.5074, -0.1278, "Europe/London"},
		{"Lisbon", 38.7223, -9.1393, "Europe/Lisbon"},
		{"Moscow", 55.7558, 37.6173, "Europe/Moscow"},
		{"Istanbul", 41.0082, 28.9784, "Europe/Istanbul"},
		{"Kaliningrad", 54.7104, 20.4522, "Europe/Kaliningrad"},
		{"Reykjavík", 64.1466, -21.9426, "Atlantic/Reykjavik"},
		{"New York", 40.7128, -74.0060, "America/New_York"},
		{"Chicago", 41.8781, -87.6298, "America/Chicago"},
		{"Phoenix", 33.4484, -112.0740, "America/Phoenix"},
		{"San Francisco", 37.7833, -122.4167, "America/Los_Angeles"},
		{"Anchorage", 61.2181, -149.9003, "America/Anchorage"},
		{"Honolulu", 21.3069, -157.8583, "Pacific/Honolulu"},
		{"Mexico City", 19.4326, -99.1332, "America/Mexico_City"},
		{"Bogotá", 4.7110, -74.0721, "America/Bogota"},
		{"São Paulo", -23.5505, -46.6333, "America/Sao_Paulo"},
		{"Nuuk", 64.1814, -51.6941, "America/Nuuk"},
		{"Lagos", 6.5244, 3.3792, "Africa/Lagos"},
		{"Cairo", 30.0444, 31.2357, "Africa/Cairo"},
		{"Nairobi", -1.2921, 36.8219, "Africa/Nairobi"},
		{"Johannesburg", -26.2041, 28.0473, "Africa/Johannesburg"},
		{"Maseru", -29.3100, 27.4783, "Africa/Maseru"}, // enclave inside South Africa
		{"Jerusalem", 31.7683, 35.2137, "Asia/Jerusalem"},
		{"Gaza", 31.5017, 34.4668, "Asia/Gaza"},
		{"Tehran", 35.6892, 51.3890, "Asia/Tehran"},
		{"Dubai", 25.2048, 55.2708, "Asia/Dubai"},
		{"Kathmandu", 27.7172, 85.3240, "Asia/Kathmandu"},
		{"Kolkata", 22.5726, 88.3639, "Asia/Kolkata"},
		{"Shanghai", 31.2304, 121.4737, "Asia/Shanghai"},
		{"Seoul", 37.5665, 126.9780, "Asia/Seoul"},
		{"Tokyo", 35.6895, 139.6917, "Asia/Tokyo"},
		{"Brisbane", -27.4698, 153.0251, "Australia/Brisbane"},
		{"Sydney", -33.8688, 151.2093, "Australia/Sydney"},
		{"Adelaide", -34.9285, 138.6007, "Australia/Adelaide"},
		{"Eucla", -31.6774, 128.8836, "Australia/Eucla"},
		{"Lord Howe Island", -31.5533, 159.0821, "Australia/Lord_Howe"},
		{"Auckland", -36.8485, 174.7633, "Pacific/Auckland"},
		{"Chatham Islands", -43.9535, -176.5597, "Pacific/Chatham"},
		{"Kiritimati", 1.8700, -157.4000, "Pacific/Kiritimati"},
		{"Troll station, Antarctica", -72.0117, 2.5350, "Antarctica/Troll"},
		{"South Pole", -89.9900, 0.0000, "Antarctica/South_Pole"},
	}
	for _, tt := range cases {
		got := LookupZoneName(tt.lat, tt.long)
		if got == "" {
			t.Errorf("%s: LookupZoneName(%v, %v) = %q, want a zone keeping the same time as %q",
				tt.place, tt.lat, tt.long, got, tt.zone)
			continue
		}
		if !sameTime(t, got, tt.zone) {
			t.Errorf("%s: LookupZoneName(%v, %v) = %q, which does not keep the same time as %q",
				tt.place, tt.lat, tt.long, got, tt.zone)
		}
	}
}

// sameTime reports whether two zones are on the same clock all year.
func sameTime(t *testing.T, a, b string) bool {
	t.Helper()
	locA, err := time.LoadLocation(a)
	if err != nil {
		t.Errorf("LoadLocation(%q): %v", a, err)
		return false
	}
	locB, err := time.LoadLocation(b)
	if err != nil {
		t.Errorf("LoadLocation(%q): %v", b, err)
		return false
	}
	// Walk a whole year so that a difference in when (or whether) the zones
	// switch to summer time shows up, not just today's offset.
	start := time.Now().UTC().Truncate(24 * time.Hour)
	for day := 0; day < 366; day++ {
		when := start.AddDate(0, 0, day)
		_, offA := when.In(locA).Zone()
		_, offB := when.In(locB).Zone()
		if offA != offB {
			return false
		}
	}
	return true
}

// TestRepresentativeZones pins the documented surprise: because zones keeping
// identical time are merged, the name you get back is often not the local one.
// These will need updating whenever the tables are regenerated, which is the
// point — the change should be visible in a diff rather than silently shipped.
func TestRepresentativeZones(t *testing.T) {
	cases := []struct {
		place     string
		lat, long float64
		want      string
	}{
		{"Berlin", 52.5200, 13.4050, "Europe/Paris"},             // all of CET-with-EU-DST
		{"Kaliningrad", 54.7104, 20.4522, "Africa/Johannesburg"}, // UTC+2, no DST
		{"Reykjavík", 64.1466, -21.9426, "Africa/Abidjan"},       // UTC+0, no DST
		{"San Francisco", 37.7833, -122.4167, "America/Los_Angeles"},
	}
	for _, tt := range cases {
		if got := LookupZoneName(tt.lat, tt.long); got != tt.want {
			t.Errorf("%s: LookupZoneName(%v, %v) = %q; want %q",
				tt.place, tt.lat, tt.long, got, tt.want)
		}
	}
}

// TestOceansAreCovered checks the other half of the dataset choice: the tables
// include the nautical Etc/GMT±N zones, so open water resolves too.
func TestOceansAreCovered(t *testing.T) {
	cases := []struct {
		where     string
		lat, long float64
	}{
		{"mid-Atlantic", 30, -40},
		{"mid-Pacific", 0, -150},
		{"Indian Ocean", -20, 80},
		{"Arctic Ocean", 85, 30},
		{"Southern Ocean", -60, -100},
		{"North Pole", 89.99, 0},
	}
	for _, tt := range cases {
		got := LookupZoneName(tt.lat, tt.long)
		if got == "" {
			t.Errorf("%s: LookupZoneName(%v, %v) = %q; want a nautical zone",
				tt.where, tt.lat, tt.long, got)
			continue
		}
		if _, err := time.LoadLocation(got); err != nil {
			t.Errorf("%s: LookupZoneName(%v, %v) = %q, which does not load: %v",
				tt.where, tt.lat, tt.long, got, err)
		}
	}
}

// TestLookupSweep walks the whole globe at a coarse resolution and checks that
// every answer is a name the time package actually knows.
func TestLookupSweep(t *testing.T) {
	loadable := map[string]bool{}
	empty := 0
	total := 0
	for lat := -89.5; lat < 90; lat += 0.5 {
		for long := -179.5; long < 180; long += 0.5 {
			total++
			zone := LookupZoneName(lat, long)
			if zone == "" {
				empty++
				continue
			}
			if _, seen := loadable[zone]; !seen {
				_, err := time.LoadLocation(zone)
				loadable[zone] = err == nil
				if err != nil {
					t.Errorf("LookupZoneName(%v, %v) = %q, which does not load: %v", lat, long, zone, err)
				}
			}
		}
	}
	if empty != 0 {
		t.Errorf("%d of %d sampled coordinates resolved to no zone; the tables cover the oceans, so none should", empty, total)
	}
	t.Logf("%d coordinates sampled, %d distinct zones returned", total, len(loadable))
}

// TestBoundsAreClamped checks the edges of the coordinate space, including
// values outside it, since callers pass through whatever a GPS gave them.
func TestBoundsAreClamped(t *testing.T) {
	cases := []struct{ lat, long float64 }{
		{90, 180}, {90, -180}, {-90, 180}, {-90, -180},
		{0, 180}, {0, -180}, {90, 0}, {-90, 0},
		{91, 181}, {-91, -181}, // out of range: must clamp, not panic
	}
	for _, tt := range cases {
		if got := LookupZoneName(tt.lat, tt.long); got == "" {
			t.Errorf("LookupZoneName(%v, %v) = %q; want a zone", tt.lat, tt.long, got)
		}
	}
}

// TestLeafTypes checks that all three encodings of the compressed tables are
// present and resolve, so a change to the generator cannot quietly stop
// emitting one of them.
func TestLeafTypes(t *testing.T) {
	unpackOnce.Do(unpackTables)

	var static, oneBit, pixmaps int
	for i, l := range leaf {
		switch l.(type) {
		case staticZone:
			static++
		case oneBitTile:
			oneBit++
		case pixmap:
			pixmaps++
		default:
			t.Fatalf("leaf[%d] has unexpected type %T", i, l)
		}
	}
	if static == 0 || oneBit == 0 || pixmaps == 0 {
		t.Errorf("leaf types: %d static zones, %d two-zone tiles, %d pixmaps; want all three to occur",
			static, oneBit, pixmaps)
	}
	t.Logf("%d leaves: %d static zones, %d two-zone tiles, %d pixmaps", len(leaf), static, oneBit, pixmaps)

	// Every mixed tile should resolve to more than one zone across its 8x8
	// pixels — that is the only reason it exists rather than a static zone.
	checked := map[string]bool{}
	for _, tl := range zoomLevels[0].tiles {
		var kind string
		switch leaf[tl.idx].(type) {
		case oneBitTile:
			kind = "two-zone tile"
		case pixmap:
			kind = "pixmap"
		default:
			continue
		}
		if checked[kind] {
			continue
		}
		checked[kind] = true

		x0, y0 := int(tl.key.X())*8, int(tl.key.Y())*8
		zones := map[string]bool{}
		for y := y0; y < y0+8; y++ {
			for x := x0; x < x0+8; x++ {
				zones[lookupPixel(x, y)] = true
			}
		}
		if len(zones) < 2 {
			t.Errorf("%s at pixel (%d,%d) resolves to %d zone(s): %v; want at least 2",
				kind, x0, y0, len(zones), zones)
		}
	}
	if len(checked) != 2 {
		t.Errorf("only found %d of the 2 mixed leaf kinds at zoom level 0", len(checked))
	}
}

func BenchmarkLookupZoneName(b *testing.B) {
	// A spread of coordinates, so the benchmark measures a mix of the big
	// single-zone tiles and the expensive per-pixel ones rather than one
	// lucky cache line.
	coords := [][2]float64{
		{52.3759, 9.7320},    // Hannover
		{37.7833, -122.4167}, // San Francisco
		{-33.8688, 151.2093}, // Sydney
		{35.6895, 139.6917},  // Tokyo
		{30, -40},            // mid-Atlantic
		{-72.0117, 2.5350},   // Antarctica
	}
	LookupZoneName(0, 0) // unpack the tables outside the timed loop
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := coords[i%len(coords)]
		LookupZoneName(c[0], c[1])
	}
}

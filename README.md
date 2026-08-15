# geotz

`geotz` maps a latitude and longitude to a timezone name — offline, in about
30 nanoseconds, with no allocations and no files to load.

```go
import "github.com/Graphmasters/geotz"

zone := geotz.LookupZoneName(52.3759, 9.7320) // Hannover
loc, err := time.LoadLocation(zone)
```

The timezone boundaries are compiled into the package as a lossy raster, so
there is nothing to download at startup and nothing to keep in sync at runtime.
That trade buys a small binary and very fast lookups at the cost of exactness
within a few kilometres of a border.

This is a fork of **[bradfitz/latlong](https://github.com/bradfitz/latlong)** by
[Brad Fitzpatrick](https://github.com/bradfitz) — thank you. The lookup
algorithm and the compressed table format are his design; this fork replaces the
data behind them and keeps the machinery working. See
[What changed](#what-changed-in-this-fork).

## What you get back

The tables are built from the **"with oceans, now"** variant of
[timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder).
Two things follow from that, and both are deliberate.

**Oceans are covered.** Every coordinate on the globe resolves to a zone. Points
at sea return a nautical `Etc/GMT±N` name rather than the empty string.

**Zones that keep the same time are merged.** Upstream's "now" dataset collapses
every zone that follows identical rules from today onward into a single region,
labelled with one representative name. There are 64 such regions, where the full
dataset has 444 zones. So:

| Coordinate | Returns | The place is really in |
|---|---|---|
| 52.52, 13.40 (Berlin) | `Europe/Paris` | `Europe/Berlin` |
| 54.71, 20.45 (Kaliningrad) | `Africa/Johannesburg` | `Europe/Kaliningrad` |
| 64.15, -21.94 (Reykjavík) | `Africa/Abidjan` | `Atlantic/Reykjavik` |
| 37.78, -122.42 (San Francisco) | `America/Los_Angeles` | `America/Los_Angeles` |

Those answers are *correct for telling the time*: `Europe/Paris` and
`Europe/Berlin` have the same offset and the same summer-time transitions, today
and for as long as the tzdata rules say they will. The test suite checks exactly
that property for a list of cities, day by day across a full year.

They are **not** correct as a name for the place, and they are **not** safe for
timestamps in the past — Berlin and Paris disagree about 1945. If you need
either, regenerate the tables from the comprehensive dataset
(`timezones-with-oceans.shapefile.zip`); the generator takes the shapefile as a
flag and nothing else has to change.

## Accuracy and cost

Measured on an Apple M1 Pro with Go 1.26, tables from
timezone-boundary-builder 2026c:

| | |
|---|---|
| Lookup | ~34 ns, 0 allocations |
| Added to your binary | ~590 KB |
| Heap after the tables unpack | ~880 KB |
| Grid resolution | 32 pixels per degree — a pixel is ~3.5 km at the equator |
| Zones in the tables | 64 |

Accuracy is exact against the source polygons at grid resolution: the
`TestAllPixels` check verifies all 66,355,152 covered pixels of the 11520×5760
grid against the rasterised shapefile, and passes with zero mismatches. Within
about a pixel of a border, the answer is whichever zone covered the majority of
that pixel.

The remaining 48 pixels of the grid — 0.00007% — are slivers where no upstream
polygon covers the raster at all, scattered along coastlines and along a few
inland borders. 42 of them survive into the tables and return the empty string;
none is in open water.

## How it works

The world is rasterised into an 11520×5760 image where each pixel's colour is
its timezone. That image is then compressed into a quadtree-ish set of six zoom
levels: a region that is a single timezone for 256×256 pixels is stored once, as
one entry; only where zones actually meet does the encoding descend to 8×8 tiles,
stored either as a one-bit bitmap between two zones or as a 4-bit-per-pixel
palette index. Identical tiles are deduplicated. The result is 17,229 leaves for
the entire planet, gzipped and base64'd into `z_gen_tables.go`.

A lookup converts the coordinate to a pixel, then probes the zoom levels from
coarsest to finest until a tile matches — usually the first one, which is why
lookups are fast and allocation-free.

## Regenerating the tables

```bash
make world              # fetch the pinned upstream release (~39 MB) into world/
make z_gen_tables.go    # rasterise it and rewrite the tables
make test-all-pixels    # verify every pixel against the shapefile
```

To move to a newer upstream release, bump `TZ_RELEASE` in the `Makefile` and run
the above. To use a different variant, change `TZ_ZIP`/`TZ_SHP` too — the
generator takes `-shapefile`, `-zone_field`, `-scale` and `-tz_release` flags, so
nothing in the code needs editing. `-write_image` drops a `regions.png` of the
rasterised world next to the tables, which is the fastest way to eyeball whether
a new dataset rendered sanely.

The generator lives behind the `geotz_gen` build tag so that ordinary builds and
tests need no dependencies beyond the standard library.

## What changed in this fork

Upstream's data source, [efele.net's `tz_world`](http://efele.net/maps/tz/world/),
is no longer maintained: its last data update was 2016-05-28 (TZ release 2016d),
and the page now carries a banner pointing at timezone-boundary-builder as its
replacement. The checked-in tables were older still — generated from 2015g
tzdata. Everything else follows from replacing that.

**Data**

- Data source switched from `tz_world` to
  [timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder),
  pinned to release **2026c** (2026-07-11) in the `Makefile`.
- Tables regenerated from that release, replacing tables built from 2015g tzdata.
- Dataset variant is "with oceans, now": sea coordinates resolve, and zones with
  identical present-day rules are merged. See [What you get back](#what-you-get-back).

**Generator correctness**

- Multi-ring polygons are handled properly. The old generator flattened every
  ring of a shape into one polygon, which was survivable for `tz_world` — hence
  the three hand-written patch polygons that used to paper over rendering
  glitches near Rome, Boise and Denver — but would have been catastrophic here,
  where a single record holds thousands of rings. Each ring is now added to the
  path separately and filled in one pass, so holes stay holes. The patch
  polygons are gone.
- Coordinates keep their sub-pixel precision instead of being truncated to whole
  pixels, so borders land on the correct side of a pixel.
- The `tzid` attribute is looked up by name rather than by column index, and the
  shapefile reader's error is checked after iteration.
- One rasteriser is reused across shapes rather than allocating a full-size one
  per polygon.

**Packaging**

- The repository is now a Go module, `github.com/Graphmasters/geotz`.
- Renamed from `latlong` to `geotz`, package and all. `LookupZoneName` keeps its
  signature.
- `io/ioutil` replaced with `io`/`os`; `// +build` replaced with `//go:build`;
  the generator's build tag is now `geotz_gen`.
- `Makefile` fetches with `curl`, pins the upstream release, and gained `world`,
  `test` and `test-all-pixels` targets.

**Tests**

- The old fixtures pinned specific pixels of the 2015 tables and could only rot.
  Tests now assert the property that matters — that the returned zone keeps the
  same time as the zone a city is really in, checked daily across a year — plus
  ocean coverage, a full-globe sweep for names the `time` package doesn't know,
  coordinate clamping at the poles and antimeridian, and that all three table
  encodings are present and resolve.
- Added `BenchmarkLookupZoneName`.

**Licensing**

- `tz_world` was public domain; timezone-boundary-builder's data is derived from
  OpenStreetMap and carries the ODbL. The obligations that come with that are
  documented in [`LICENSES.md`](LICENSES.md), with attribution in
  [`NOTICE`](NOTICE), the licence text in [`DATA_LICENSE`](DATA_LICENSE), and a
  provenance header in the generated file.

What did *not* change: the lookup algorithm, the table format, and the public
API are all still Brad Fitzpatrick's.

## Licences

The code is Apache-2.0 ([`COPYING`](COPYING)). The timezone boundary data
compiled into `z_gen_tables.go` is ODbL v1.0 ([`DATA_LICENSE`](DATA_LICENSE)):

> Contains information from OpenStreetMap, which is made available under the
> Open Database License (ODbL) v1.0. © OpenStreetMap contributors —
> <https://www.openstreetmap.org/copyright>
>
> Boundaries derived via
> [timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder),
> release 2026c.

**[`LICENSES.md`](LICENSES.md) explains what that means in practice** — for
calling the API, for shipping a binary, for running a public service, and for
internal use — and why the generated tables are treated as a derivative database.

## Acknowledgements

- **[Brad Fitzpatrick](https://github.com/bradfitz)** for
  [latlong](https://github.com/bradfitz/latlong), which this is a fork of: the
  rasterise-and-compress idea, the tile encoding, and the lookup code.
- **[Evan Siroky](https://github.com/evansiroky)** for
  [timezone-boundary-builder](https://github.com/evansiroky/timezone-boundary-builder),
  which keeps the boundaries current.
- **Eric Muller** for `tz_world`, which served this library — and most of the
  ecosystem — for over a decade.
- **OpenStreetMap contributors**, whose mapping the boundaries are built from.

## A note on how this fork was made

The changes described above were made by [Claude
Code](https://claude.com/claude-code), working from a short brief: credit the
upstream author, move off the unmaintained data source, evaluate the licence
obligations that come with the new one, and document all of it. The data
migration, the generator fixes, the tests and the licence documentation are its
work, reviewed by a human before merging.

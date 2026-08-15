# Licences

This repository contains two things under two different licences: **source
code** under the Apache License 2.0, and **timezone boundary data** under the
Open Database License. Both ship in the same Go module, and importing the
package pulls in both.

> This document is an engineering assessment of what the licences require, not
> legal advice. Where a question is genuinely unsettled, it says so rather than
> pretending otherwise.

## What is licensed how

| Path | Licence | Copyright |
|---|---|---|
| `geotz.go`, `geotz_test.go`, `gen_test.go`, `Makefile`, docs | Apache License 2.0 — see [`COPYING`](COPYING) | 2014 Google Inc., and later contributors |
| `z_gen_tables.go` — *the data encoded in it* | Open Database License (ODbL) v1.0 — see [`DATA_LICENSE`](DATA_LICENSE) | © OpenStreetMap contributors |

`z_gen_tables.go` is generated, so it has two aspects. The Go syntax wrapped
around the tables comes from `gen_test.go` and is Apache-2.0 like the rest of
the code; the bytes inside the string literals are compressed OpenStreetMap
boundary data and are ODbL. The file carries a header saying so.

ODbL §2 makes this split work: *"This License does not apply to computer
programs used in the making or operation of the Database."* The licences govern
different subject matter, so there is no copyleft bridge from the data to your
application code.

## Where the data comes from

```
OpenStreetMap  →  timezone-boundary-builder 2026c  →  rasterised tables in z_gen_tables.go
   (ODbL)            (MIT code, ODbL output)              (this repository)
```

The exact input is the `timezones-with-oceans-now.shapefile.zip` asset of
[timezone-boundary-builder release
2026c](https://github.com/evansiroky/timezone-boundary-builder/releases/tag/2026c),
pinned in the [`Makefile`](Makefile). `make z_gen_tables.go` downloads that file
and regenerates the tables; `gen_test.go` is the complete algorithm that turns
one into the other.

Before this fork, the data came from Eric Muller's `tz_world` at
[efele.net](http://efele.net/maps/tz/world/), which Muller placed in the public
domain (CC0) and which carried no attribution or share-alike obligation at all.
That dataset stopped being updated on 2016-05-28 (TZ release 2016d) and its own
page now points at timezone-boundary-builder as its successor. **Switching data
sources is what brought the obligations below into this repository** — they did
not exist while the tables were built from `tz_world`.

## Attribution

Anywhere you would normally state licences — a README, an about screen, a
`NOTICE` file, API documentation — reproduce this:

```
Timezone boundary data © OpenStreetMap contributors, available under the
Open Database License (ODbL) v1.0: https://opendatacommons.org/licenses/odbl/1-0/
Derived via https://github.com/evansiroky/timezone-boundary-builder (release 2026c)
```

The OSM Foundation's [Attribution
Guidelines](https://osmfoundation.org/wiki/Licence/Attribution_Guidelines) ask
for credit to OpenStreetMap plus either the ODbL text or a link to it, placed
"in a location (such as a relevant directory) where users would be likely to
look for it, such as a readme file, or within the data or metadata". For a
library that means the README and the generated file, both of which carry it
here.

## What this means for you

**If you call `LookupZoneName` and use the answer**, nothing follows. A single
coordinate → zone answer is an insubstantial extract; the OSMF [Geocoding
Guideline](https://osmfoundation.org/wiki/Licence/Community_Guidelines/Geocoding_-_Guideline)
treats individual geocoding results that way and says they do not trigger
share-alike. Attribution is still expected of an application that offers such
lookups to the public.

**If you ship a binary that imports this package**, you are distributing the
tables. Keep the attribution above with your product, don't relicense the data,
and don't wrap it in terms that restrict what recipients may do with it
(ODbL §4.7).

**If you run a public service on top of it**, note that ODbL §4.4 and §4.6 key
on *Publicly Use*, not on shipping bytes. A hosted API that never distributes
the tables but answers lookups publicly is still publicly using a derivative
database, and owes the same share-alike and access obligations. The Geocoding
Guideline states this directly for geocoders built on a derivative database.

**If you use it purely inside your own organisation**, ODbL §4.5 exempts you —
internal use is not public use, and none of this applies.

**If you modify the data or regenerate the tables from a different source**,
that new table is yours to publish under the ODbL, with your own provenance
statement replacing the one in the generated header.

## How this repository meets the obligations

| ODbL | Requirement | How it is met here |
|---|---|---|
| §4.2 Notices | Convey only under this licence; include the licence or its URI, in the database and in the documentation; keep existing notices intact | [`DATA_LICENSE`](DATA_LICENSE) holds the full text; the header of `z_gen_tables.go` and the package doc carry the notice; this file and the README repeat it. The upstream shapefile zips contain no notice file, so there was nothing to keep intact — this one originates it |
| §4.3 Notice for produced works | Make users aware the content comes from the database and is available under this licence | Same notice, plus the attribution string above for downstream products |
| §4.4 Share alike | A publicly used derivative database must be under this licence, a later version, or a compatible one | The data in `z_gen_tables.go` is ODbL v1.0. It is *not* covered by `COPYING` |
| §4.6 Access to derivative databases | Offer either the whole derivative database *or* a file of the alterations *or* the method of making them, "such as an algorithm" | This repository takes the third option and publishes the method: `gen_test.go` plus a `Makefile` that names the exact upstream release. The generated tables themselves are also right here in the repository |
| §4.7 Technological measures | No DRM or added terms restricting the licence | None applied |
| §4.8 Licensing of others | No sublicensing | Nothing here purports to sublicense the data; recipients get their licence from the ODbL directly |
| Apache-2.0 §4 | Retain notices, state changes | `COPYING` and the original `Copyright 2014 Google Inc.` headers are intact; changes are listed in the README and in `NOTICE` |

## The judgment call

ODbL treats a *Derivative Database* (share-alike applies) differently from a
*Produced Work* (it does not). Our compressed lookup table can be argued either
way, and the OSM Foundation has published nothing that settles it:

- **Produced Work.** The table literally is a raster image, and the OSMF
  [Produced Work
  Guideline](https://osmfoundation.org/wiki/Licence/Community_Guidelines/Produced_Work_-_Guideline)
  lists "any raster image" among things that are usually produced works.
- **Derivative Database.** The same guideline says that *"if the published
  result of your project is intended for the extraction of the original data,
  then it is a database and not a Produced Work"* — and extracting the original
  data is the only thing this table is for. ODbL §1.0's definition of a
  derivative database ("any translation, adaptation, arrangement, modification,
  or any other alteration") is deliberately broad, and the whole dataset goes
  in, so the extract is comfortably "substantial".

**This repository takes the conservative route and treats the tables as a
derivative database.** It costs almost nothing — attribution, a licence file, a
pointer to the generator — and it avoids relying on a classification that
nobody has ruled on. Note also that even under the produced-work reading, the
OSMF guideline says the underlying database or the alterations still have to be
published, which is what §4.6 above does anyway.

Most comparable projects (timeshape, timezonefinder, ugjka/go-tz,
ringsaturn/tzf) split their licences the same way, with a `DATA_LICENSE`
alongside the code licence. Few of them carry the OpenStreetMap attribution
itself; this repository does.

## If you would rather not take on the data licence

The obligations travel with the tables, not with the algorithm. If a consumer
needs Apache-2.0-only bytes, the options are to keep the lookup code and
regenerate `z_gen_tables.go` from a public-domain or permissively licensed
boundary set, or to split the tables into a separate module so the licence
boundary matches the import boundary. Neither is done here — this module ships
both.

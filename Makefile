# The timezone-boundary-builder release the checked-in tables are built from.
# Releases are named after the tzdata release they follow, e.g. 2026c.
TZ_RELEASE = 2026c

# Which variant of the release to use. "with-oceans" so that points at sea
# resolve to a nautical Etc/GMT±N zone instead of nothing, "now" so that zones
# following identical present-day rules are merged into one region. See
# README.md for what that means for the names this package returns.
TZ_ZIP = timezones-with-oceans-now.shapefile.zip
TZ_SHP = world/combined-shapefile-with-oceans-now.shp

TZ_URL = https://github.com/evansiroky/timezone-boundary-builder/releases/download/$(TZ_RELEASE)/$(TZ_ZIP)

# The generator is a separate module (see gen/go.mod), so it is run from its own
# directory and writes back up here.
.PHONY: z_gen_tables.go
z_gen_tables.go: gen/main.go gen/tables.go internal/tile/tile.go $(TZ_SHP)
	cd gen && go run . -shapefile ../$(TZ_SHP) -tz_release $(TZ_RELEASE) -o ../z_gen_tables.go

# make world downloads and unpacks the boundary shapefile (~39 MB download,
# ~60 MB unpacked) into world/, which is .gitignore'd.
.PHONY: world
world: $(TZ_SHP)

$(TZ_SHP):
	mkdir -p world
	curl -fSL -o world/$(TZ_ZIP) $(TZ_URL)
	unzip -o -d world world/$(TZ_ZIP)
	rm world/$(TZ_ZIP)

.PHONY: test
test:
	go test ./...
	cd gen && go test ./...

# The exhaustive pixel-by-pixel check of the generated tables against the
# rasterized shapefile. Needs the shapefile; skips without it.
.PHONY: test-all-pixels
test-all-pixels: $(TZ_SHP)
	cd gen && go test -run TestAllPixels -v -timeout 30m ./...

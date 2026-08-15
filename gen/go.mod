// The table generator is its own module so that its dependencies — a shapefile
// reader and a rasterizer — never reach anyone who imports the package. Nothing
// consumes this module; it is only ever run from the repository, so it takes
// the parent from disk rather than from a published version.
module github.com/Graphmasters/geotz/gen

go 1.22

require (
	github.com/Graphmasters/geotz v0.0.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jonas-p/go-shp v0.1.1
	golang.org/x/image v0.18.0
)

replace github.com/Graphmasters/geotz => ../

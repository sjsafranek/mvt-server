

ALTER TABLE layers ADD COLUMN is_updatable BOOLEAN DEFAULT false;



## 0.01.11
### Changed
 - logging to files
### Fixed
 - guard against panic for null extent


## 0.01.10
### Fixed
 - panic if postgres cannot be reached


## 0.01.09
### Added
 - View layer extent in web view
### Fixed
 - fix for case sensitivity when adding layer


## 0.01.08
### Added
 - is_updatable added as a layer column, this overrides caching
### Fixed
 - fix for shp2pgsql import, writes to tmp file before importing to postgis


## 0.01.07
### Added
 - filter for cook tiles
 - bbox for cook tiles


## 0.01.06
### Added
 - new filter methods: range


## 0.01.05
### Added
 - new filter methods: within & overlaps
 - enabled go modules --> go.mod and go.sum


## 0.01.04
### Changed
 - mbtiles moved to its own package
 - refactor of tile caching
### Fixed
 - mbtiles multiple metadata inserts
 - context nil error crash


## 0.01.03
### Added
 - mbtiles database for caching


## 0.01.02
### Added
 - python unittests for api


## 0.01.01
### Changed
 - simplified GeoDatabase query pipeline
 - cleanup


## 0.01.00
### Added
 - added context for cancellation of tiles
### Changed
 - Updated logging to handling cancellation messages
 - api layers and layer return from cache instead of database


## 0.00.07
### Added
 - tcp command port (delete layer, upload layer)
### Changed
 - layer collection talks to database directly
 - layer collection holds layers that talk to database
 - layer fetches tile from database
### Fixed
 - bug with sql scan json_build_object to interface


## 0.00.06
### Added
 - lazy loading of layers
### Changed
 - layers struct has add and delete layer


## 0.00.05
### Added
 - added CORS middleware
 - Redis support for tile caching
### Changed
 - moved Web Feature Service (wfs) to different route for easier layer checking
### Fixed
 - OPTIONS request failing, added CORS middleware


## 0.00.04
### Added
 - Web Feature Service (wfs) for querying layers
 - Tile caching options added to config


## 0.00.03
### Added
 - added srid column to layers table
 - tile cooking for cache
 - docker-compose.yml for postgis database
### Changed
 - Database query work flow
### Fixed
 - config changes for connecting to remote databases
 - more memory efficient database queries for SRID transformations


## 0.00.02
### Added
 - config for connecting to postgres
 - database tables and functions get created if not exists
 - uploading shapefiles


## 0.00.01
### Added
 - mapbox vector tile creation in postgres
 - layers table

package tilecache

import (
	"errors"

	"mvt-server/lib/tilecache/nocache"
	"mvt-server/lib/tilecache/disk"
	"mvt-server/lib/tilecache/memory"
	"mvt-server/lib/tilecache/mbtile"
	"mvt-server/lib/tilecache/sqlite"
	"mvt-server/lib/tilecache/redis"
)


func NewTileCache(options Config) (TileCache, error) {
	if "" == options.Type {
		options.Type = "disk"
	}

	switch options.Type {
	case "none":
		return nocache.New()
	case "memory":
		return memory.New()
	case "disk":
		return disk.New(options.Directory)
	case "mbtiles":
		return mbtiles.New(options.Directory)
	case "sqlite3":
		return sqlite.New(options.Directory)
	case "redis":
		if 0 == options.Port {
			options.Port = REDIS_PORT
		}
		return redis.New(options.Port)
	default:
		return nil, errors.New("Unknown cache type")
	}
}

type TileCache interface {
	GetTile(string, uint32, uint32, uint32) ([]uint8, error)
	SetTile(string, uint32, uint32, uint32, []uint8) error
	SetMetadata(string, [][]string) error
}

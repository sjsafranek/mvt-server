package tilecache

const (
	DEFAULT_TILE_CACHE_DIRECTORY string = "cache"
	DEFAULT_TILE_CACHE_TYPE      string = "disk"
	DEFAULT_REDIS_PORT           int    = 6379
)

var (
	TILE_CACHE_DIRECTORY = DEFAULT_TILE_CACHE_DIRECTORY
	TILE_CACHE_TYPE      = DEFAULT_TILE_CACHE_TYPE
	REDIS_PORT           = DEFAULT_REDIS_PORT
)

type Config struct {
	Directory string `toml:"directory"`
	Type      string `toml:"type"`
	Port      int    `toml:"port"`
}

func (self *Config) GetDirectory() string {
	if "" != self.Directory {
		return self.Directory
	}
	return TILE_CACHE_DIRECTORY
}

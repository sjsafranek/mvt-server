package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

func New(port int) (*RedisTileCache, error) {
	addr := fmt.Sprintf(":%v", port)
	return &RedisTileCache{pool: &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}}, nil
}

type RedisTileCache struct {
	pool *redis.Pool
}

func (self *RedisTileCache) GetTile(layerName string, x, y, z uint32) ([]uint8, error) {
	conn := self.pool.Get()
	defer conn.Close()
	tileName := fmt.Sprintf("%v-%v-%v-%v.mvt", layerName, z, x, y)
	tileData, err := redis.Bytes(conn.Do("GET", tileName))
	return tileData, err
}

func (self *RedisTileCache) SetTile(layerName string, x, y, z uint32, tileData []uint8) error {
	conn := self.pool.Get()
	defer conn.Close()
	tileName := fmt.Sprintf("%v-%v-%v-%v.mvt", layerName, z, x, y)
	_, err := redis.String(conn.Do("SET", tileName, tileData))
	return err
}

func (self *RedisTileCache) SetMetadata(layerName string, metadata [][]string) error {
	return nil
}

package cache

import "github.com/redis/go-redis/v9"

type FavoriteCache struct {
	cmd redis.Cmdable
}

func NewFavoriteCache(cmd redis.Cmdable) *FavoriteCache {
	return &FavoriteCache{cmd: cmd}
}

package repository

import (
	"github.com/crazyfrankie/favorite/biz/repository/cache"
	"github.com/crazyfrankie/favorite/biz/repository/dao"
)

type FavoriteRepo struct {
	cache *cache.FavoriteCache
	dao   *dao.FavoriteDao
}

func NewFavoriteRepo(c *cache.FavoriteCache, d *dao.FavoriteDao) *FavoriteRepo {
	return &FavoriteRepo{
		cache: c,
		dao:   d,
	}
}

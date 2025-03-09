package repository

import (
	"context"

	"github.com/crazyfrankie/favorite/internal/biz/repository/cache"
	"github.com/crazyfrankie/favorite/internal/biz/repository/dao"
)

var (
	ErrAlreadyExists = cache.ErrAlreadyExists
	ErrNotFound      = cache.ErrNotFound
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

// CreateFavorite 创建点赞记录及递增点赞数
func (r *FavoriteRepo) CreateFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	return r.cache.CreateFavorite(ctx, biz, bizId, uid)
}

// DeleteFavorite 删除点赞记录及递减点赞数
func (r *FavoriteRepo) DeleteFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	return r.cache.DeleteFavorite(ctx, biz, bizId, uid)
}

// FavoriteCount 获取单个内容的点赞总数
func (r *FavoriteRepo) FavoriteCount(ctx context.Context, biz string, bizId int64) (int64, error) {
	return r.cache.FavoriteCount(ctx, biz, bizId)
}

// BizFavoriteUser 获取某个内容的点赞用户
func (r *FavoriteRepo) BizFavoriteUser(ctx context.Context, biz string, bizId int64) ([]int64, error) {
	return r.cache.BizFavoriteUser(ctx, biz, bizId)
}

// UserFavoriteCount 获取用户的点赞内容总数
func (r *FavoriteRepo) UserFavoriteCount(ctx context.Context, uid int64) (int64, error) {
	return r.cache.UserFavoriteCount(ctx, uid)
}

// UserFavoriteElements 用户点赞的内容 ID 集合
func (r *FavoriteRepo) UserFavoriteElements(ctx context.Context, uid int64) ([]int64, error) {
	return r.cache.UserFavoriteElements(ctx, uid)
}

// UserFavoritedCount 获取用户的内容被点赞总数
func (r *FavoriteRepo) UserFavoritedCount(ctx context.Context, biz string, bizIds []int64) (int64, error) {
	return r.cache.UserFavoritedCount(ctx, biz, bizIds)
}

// IsUserFavorite 用户是否点赞了某个内容
func (r *FavoriteRepo) IsUserFavorite(ctx context.Context, uid, bizId int64) (bool, error) {
	return r.cache.IsUserFavorite(ctx, uid, bizId)
}

// BatchIsUserFavorite 批量查询用户是否点赞了某个内容
func (r *FavoriteRepo) BatchIsUserFavorite(ctx context.Context, uid int64, bizIds []int64) (map[int64]bool, error) {
	return r.cache.BatchIsUserFavorite(ctx, uid, bizIds)
}

// GetTopFavoriteContent 点赞数排行榜
func (r *FavoriteRepo) GetTopFavoriteContent(ctx context.Context, biz string, topN int64) ([]int64, error) {
	return r.cache.GetTopFavoriteContent(ctx, biz, topN)
}

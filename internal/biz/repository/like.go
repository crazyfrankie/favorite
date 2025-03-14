package repository

import (
	"context"

	"github.com/crazyfrankie/favorite/internal/biz/domain"
	"github.com/crazyfrankie/favorite/internal/biz/repository/cache"
	"github.com/crazyfrankie/favorite/internal/biz/repository/dao"
)

var (
	ErrAlreadyExists = cache.ErrAlreadyExists
	ErrNotFound      = cache.ErrNotFound
)

type FavoriteRepo struct {
	cache *cache.FavoriteCache
	write *dao.FavoriteWriteDao
	read  *dao.FavoriteReadDao
}

func NewFavoriteRepo(c *cache.FavoriteCache, write *dao.FavoriteWriteDao, read *dao.FavoriteReadDao) *FavoriteRepo {
	return &FavoriteRepo{
		cache: c,
		write: write,
		read:  read,
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
func (r *FavoriteRepo) UserFavoriteElements(ctx context.Context, uid int64) ([]string, error) {
	return r.cache.UserFavoriteElements(ctx, uid)
}

// UserFavoritedCount 获取用户的内容被点赞总数
func (r *FavoriteRepo) UserFavoritedCount(ctx context.Context, biz string, bizIds []int64) (int64, error) {
	return r.cache.UserFavoritedCount(ctx, biz, bizIds)
}

// IsUserFavorite 用户是否点赞了某个内容
func (r *FavoriteRepo) IsUserFavorite(ctx context.Context, biz string, uid, bizId int64) (bool, error) {
	return r.cache.IsUserFavorite(ctx, biz, uid, bizId)
}

// GetTopFavoriteContent 点赞数排行榜
func (r *FavoriteRepo) GetTopFavoriteContent(ctx context.Context, biz string, topN int64) ([]int64, error) {
	return r.cache.GetTopFavoriteContent(ctx, biz, topN)
}

// SyncFavoritesCount 将内容点赞总数同步到数据库
func (r *FavoriteRepo) SyncFavoritesCount(ctx context.Context) error {
	countStream, err := r.cache.GetAllCount(ctx)
	if err != nil {
		return err
	}

	// 设置批量提交大小，避免频繁写入
	batchSize := 50
	var batch []domain.FavoriteCount

	// 持续消费 channel
	for count := range countStream {
		batch = append(batch, count)

		// 如果达到批量大小, 就进行批量写入
		if len(batch) >= batchSize {
			if err := r.write.SaveFavoriteCounts(ctx, batch); err != nil {
				return err
			}
			// 清空 batch
			batch = batch[:0]
		}
	}

	// 处理最后剩余的数据
	if len(batch) > 0 {
		if err := r.write.SaveFavoriteCounts(ctx, batch); err != nil {
			return err
		}
	}

	return nil
}

package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyExists = errors.New("favorite already exists")
	ErrNotFound      = errors.New("favorite not found")
)

type FavoriteCache struct {
	cmd redis.Cmdable
}

func NewFavoriteCache(cmd redis.Cmdable) *FavoriteCache {
	return &FavoriteCache{cmd: cmd}
}

// CreateFavorite 创建点赞记录及递增点赞数
func (c *FavoriteCache) CreateFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	userKey := c.userKey(uid)
	countKey := c.countKey(biz)
	field := fmt.Sprintf("%d", bizId)

	// 先判断是否已点赞（幂等性检查）
	exists, err := c.cmd.SIsMember(ctx, userKey, bizId).Result()
	if err != nil {
		return err
	}
	if exists { // 如果已经点过赞，直接返回错误，防止重复点赞
		return ErrAlreadyExists
	}

	// 事务性增加点赞记录 + 递增点赞数
	pipe := c.cmd.TxPipeline()
	pipe.SAdd(ctx, userKey, bizId)        // 记录点赞
	pipe.HIncrBy(ctx, countKey, field, 1) // 点赞数+1
	_, err = pipe.Exec(ctx)               // 执行事务

	return err
}

// DeleteFavorite 删除点赞记录及递减点赞数
func (c *FavoriteCache) DeleteFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	userKey := c.userKey(uid)
	countKey := c.countKey(biz)
	field := fmt.Sprintf("%d", bizId)

	// 先判断是否已点赞（幂等性检查）
	exists, err := c.cmd.SIsMember(ctx, userKey, bizId).Result()
	if err != nil {
		return err
	}
	if !exists { // 如果点赞记录不存在，直接返回错误，防止重复取消
		return ErrNotFound
	}

	// 事务性删除点赞记录 + 递减点赞数
	pipe := c.cmd.TxPipeline()
	pipe.SRem(ctx, userKey, bizId)         // 删除点赞记录
	pipe.HIncrBy(ctx, countKey, field, -1) // 点赞数-1
	_, err = pipe.Exec(ctx)                // 执行事务

	return err
}

// FavoriteCount 获取单个内容的点赞总数
func (c *FavoriteCache) FavoriteCount(ctx context.Context, biz string, bizId int64) (int64, error) {
	countKey := c.countKey(biz)
	field := fmt.Sprintf("%d", bizId)

	res, err := c.cmd.HGet(ctx, countKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil // 如果 key 不存在，返回 0
	}
	if err != nil {
		return 0, err // 其他错误返回
	}

	return res, nil
}

// UserFavoriteCount 获取用户的点赞内容总数
func (c *FavoriteCache) UserFavoriteCount(ctx context.Context, uid int64) (int64, error) {
	userKey := c.userKey(uid)

	res, err := c.cmd.SCard(ctx, userKey).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil // 用户从未点赞，返回 0
		}
		return 0, err
	}

	return int64(res), nil
}

// UserFavoriteElements 用户点赞的内容 ID 集合
func (c *FavoriteCache) UserFavoriteElements(ctx context.Context, uid int64) ([]int64, error) {
	userKey := c.userKey(uid)

	result, err := c.cmd.SDiff(ctx, userKey).Result()
	if err != nil {
		return []int64{}, err
	}

	res := make([]int64, len(result))
	for _, v := range result {
		val, _ := strconv.ParseInt(v, 10, 0)
		res = append(res, val)
	}

	return res, nil
}

// UserFavoritedCount 获取用户的内容被点赞总数
func (c *FavoriteCache) UserFavoritedCount(ctx context.Context, biz string, bizIds []int64) (int64, error) {
	countKey := c.countKey(biz)
	fields := make([]string, len(bizIds))
	for i, id := range bizIds {
		fields[i] = fmt.Sprintf("%d", id)
	}

	res, err := c.cmd.HMGet(ctx, countKey, fields...).Result()
	if err != nil {
		return 0, err
	}

	var count int64
	for _, v := range res {
		if v == nil {
			continue
		}
		val, err := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
		if err != nil {
			zap.L().Warn("Failed to parse count", zap.Any("value", v), zap.Error(err))
			continue
		}
		count += val
	}

	return count, nil
}

// IsUserFavorite 用户是否点赞了某个内容
func (c *FavoriteCache) IsUserFavorite(ctx context.Context, uid, bizId int64) (bool, error) {
	userKey := c.userKey(uid)
	exists, err := c.cmd.SIsMember(ctx, userKey, bizId).Result()
	if err != nil {
		return false, err
	}

	return exists, nil
}

// BatchIsUserFavorite 批量查询用户是否点赞了某个内容
func (c *FavoriteCache) BatchIsUserFavorite(ctx context.Context, uid int64, bizIds []int64) (map[int64]bool, error) {
	userKey := c.userKey(uid)
	args := make([]interface{}, len(bizIds))
	for i, id := range bizIds {
		args[i] = id
	}

	res, err := c.cmd.SMIsMember(ctx, userKey, args...).Result()
	if err != nil {
		return nil, err
	}

	resultMap := make(map[int64]bool, len(bizIds))
	for i, v := range res {
		resultMap[bizIds[i]] = v
	}

	return resultMap, nil
}

// GetTopFavoriteContent 点赞数排行榜
func (c *FavoriteCache) GetTopFavoriteContent(ctx context.Context, biz string, topN int64) ([]int64, error) {
	countKey := c.countKey(biz)

	// 按照点赞数降序获取前 topN 个内容
	res, err := c.cmd.ZRevRange(ctx, countKey, 0, topN-1).Result()
	if err != nil {
		return nil, err
	}

	topBizIds := make([]int64, len(res))
	for i, v := range res {
		id, _ := strconv.ParseInt(v, 10, 64)
		topBizIds[i] = id
	}

	return topBizIds, nil
}

func (c *FavoriteCache) countKey(biz string) string {
	return fmt.Sprintf("favorite:count:%s", biz)
}

func (c *FavoriteCache) userKey(uid int64) string {
	return fmt.Sprintf("favorite:user:%d", uid)
}

package cache

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/crazyfrankie/favorite/internal/biz/domain"
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

func (c *FavoriteCache) keys() struct {
	// 全局计数器hash, field为"{biz}:{bizId}", value为点赞数
	countKey string
	// 全局业务类型set，记录所有biz
	bizTypesKey string
	// 业务维度的点赞用户set模板, 填充biz,bizId后使用
	bizUserKey string
	// 用户维度的点赞记录set模板, 填充uid后使用
	userFavoriteKey   string
	userUnFavoriteKey string
} {
	return struct {
		countKey          string
		bizTypesKey       string
		bizUserKey        string
		userFavoriteKey   string
		userUnFavoriteKey string
	}{
		countKey:          "favorite:counts",          // 全局计数器
		bizTypesKey:       "favorite:biz:types",       // 业务类型集合
		bizUserKey:        "favorite:biz:%s:%d:users", // 记录内容被谁点赞
		userFavoriteKey:   "favorite:user:%d",         // 记录用户点赞了什么
		userUnFavoriteKey: "unfavorite:user:%d",       // 记录用户取消点赞了什么
	}
}

// CreateFavorite 创建点赞时维护业务类型
func (c *FavoriteCache) CreateFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	keys := c.keys()

	pipe := c.cmd.TxPipeline()
	pipe.SAdd(ctx, keys.bizTypesKey, biz)

	field := fmt.Sprintf("%s:%d", biz, bizId)
	pipe.HIncrBy(ctx, keys.countKey, field, 1)

	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)
	pipe.ZAdd(ctx, userKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: fmt.Sprintf("%s:%d", biz, bizId),
	})

	// 记录内容被点赞
	bizUserKey := fmt.Sprintf(keys.bizUserKey, biz, bizId)
	pipe.SAdd(ctx, bizUserKey, uid)
	pipe.Expire(ctx, userKey, 7*24*time.Hour)

	_, err := pipe.Exec(ctx)

	return err
}

// DeleteFavorite 删除点赞记录及递减点赞数
func (c *FavoriteCache) DeleteFavorite(ctx context.Context, biz string, bizId, uid int64) error {
	keys := c.keys()

	pipe := c.cmd.TxPipeline()
	// 更新计数
	field := fmt.Sprintf("%s:%d", biz, bizId)
	pipe.HIncrBy(ctx, keys.countKey, field, -1)

	// 记录用户取消点赞
	userKey := fmt.Sprintf(keys.userUnFavoriteKey, uid)
	pipe.ZAdd(ctx, userKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: fmt.Sprintf("%s:%d", biz, bizId),
	})

	// 删除点赞内容记录
	bizUserKey := fmt.Sprintf(keys.bizUserKey, biz, bizId)
	pipe.SRem(ctx, bizUserKey, uid)

	pipe.Expire(ctx, userKey, 7*24*time.Hour)
	_, err := pipe.Exec(ctx)

	return err
}

// FavoriteCount 获取单个内容的点赞总数
func (c *FavoriteCache) FavoriteCount(ctx context.Context, biz string, bizId int64) (int64, error) {
	keys := c.keys()

	field := fmt.Sprintf("%s:%d", biz, bizId)
	res, err := c.cmd.HGet(ctx, keys.countKey, field).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	return res, nil
}

// BizFavoriteUser 获取某个内容的点赞用户
func (c *FavoriteCache) BizFavoriteUser(ctx context.Context, biz string, bizId int64) ([]int64, error) {
	keys := c.keys()

	bizUserKey := fmt.Sprintf(keys.bizUserKey, biz, bizId)
	res, err := c.cmd.SMembers(ctx, bizUserKey).Result()
	if err != nil {
		return nil, err
	}

	users := make([]int64, len(res))
	for _, v := range res {
		uid, _ := strconv.ParseInt(v, 10, 64)
		users = append(users, uid)
	}

	return users, nil
}

// UserFavoriteCount 获取用户的点赞内容总数
func (c *FavoriteCache) UserFavoriteCount(ctx context.Context, uid int64) (int64, error) {
	keys := c.keys()

	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)
	res, err := c.cmd.SCard(ctx, userKey).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, err
	}

	return int64(res), nil
}

// UserFavoriteElements 用户点赞的内容集合
func (c *FavoriteCache) UserFavoriteElements(ctx context.Context, uid int64) ([]string, error) {
	keys := c.keys()

	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)
	res, err := c.cmd.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// UserFavoritedCount 获取用户的内容被点赞总数
func (c *FavoriteCache) UserFavoritedCount(ctx context.Context, biz string, bizIds []int64) (int64, error) {
	keys := c.keys()

	var count int64
	for _, v := range bizIds {
		bizUserKey := fmt.Sprintf(keys.bizUserKey, biz, v)
		res, err := c.cmd.SCard(ctx, bizUserKey).Result()
		if err != nil {
			return 0, nil
		}
		count += res
	}

	return count, nil
}

// IsUserFavorite 用户是否点赞了某个内容
func (c *FavoriteCache) IsUserFavorite(ctx context.Context, biz string, uid, bizId int64) (bool, error) {
	keys := c.keys()

	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)
	res, err := c.cmd.SIsMember(ctx, userKey, fmt.Sprintf("%s:%d", biz, bizId)).Result()
	if err != nil {
		return false, err
	}

	return res, nil
}

// GetTopFavoriteContent 点赞数排行榜
func (c *FavoriteCache) GetTopFavoriteContent(ctx context.Context, biz string, topN int64) ([]int64, error) {
	keys := c.keys()

	// 获取所有计数
	all, err := c.cmd.HGetAll(ctx, keys.countKey).Result()
	if err != nil {
		return nil, err
	}

	// 过滤出指定biz的数据并排序
	type pair struct {
		bizId int64
		count int64
	}
	pairs := make([]pair, 0, len(all))

	for field, count := range all {
		parts := strings.Split(field, ":")
		if len(parts) != 2 || parts[0] != biz {
			continue
		}

		bizId, _ := strconv.ParseInt(parts[1], 10, 64)
		cnt, _ := strconv.ParseInt(count, 10, 64)
		pairs = append(pairs, pair{bizId: bizId, count: cnt})
	}

	// 按点赞数排序
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})

	// 取前N个
	if int64(len(pairs)) > topN {
		pairs = pairs[:topN]
	}

	result := make([]int64, len(pairs))
	for i, p := range pairs {
		result[i] = p.bizId
	}

	return result, nil
}

// GetUserRecentFavorites 获取用户最近的点赞记录
func (c *FavoriteCache) GetUserRecentFavorites(ctx context.Context, uid int64, limit int64) ([]string, error) {
	keys := c.keys()
	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)

	// 使用 ZREVRANGE 获取最近的点赞记录
	return c.cmd.ZRevRange(ctx, userKey, 0, limit-1).Result()
}

// GetUserUnFavorites 获取用户的取消点赞记录
func (c *FavoriteCache) GetUserUnFavorites(ctx context.Context, uid int64) (map[string]int64, error) {
	keys := c.keys()
	userKey := fmt.Sprintf(keys.userUnFavoriteKey, uid)

	result, err := c.cmd.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, err
	}

	res := make(map[string]int64, len(result))
	for _, v := range result {
		strs := strings.Split(v, ":")
		bizId, _ := strconv.ParseInt(strs[1], 10, 64)
		res[strs[0]] = bizId
	}
	return res, nil
}

// CleanupUserHistory 清理用户的历史记录
func (c *FavoriteCache) CleanupUserHistory(ctx context.Context, uid int64, days int) error {
	keys := c.keys()
	deadline := time.Now().AddDate(0, 0, -days).Unix()

	userKey := fmt.Sprintf(keys.userFavoriteKey, uid)
	unFavoriteKey := fmt.Sprintf(keys.userUnFavoriteKey, uid)

	pipe := c.cmd.TxPipeline()
	pipe.ZRemRangeByScore(ctx, userKey, "0", fmt.Sprintf("%d", deadline))
	pipe.Del(ctx, unFavoriteKey)

	_, err := pipe.Exec(ctx)
	return err
}

// GetAllCount 获取全局点赞数
func (c *FavoriteCache) GetAllCount(ctx context.Context) (<-chan domain.FavoriteCount, error) {
	keys := c.keys()
	countKey := keys.countKey

	// 使用 channel 实现数据流式返回
	out := make(chan domain.FavoriteCount, 100)
	go func() {
		defer close(out)

		var cursor uint64
		for {
			// HSCAN 分批获取数据
			res, newCursor, err := c.cmd.HScan(ctx, countKey, cursor, "", 100).Result()
			if err != nil {
				// 不返回错误，避免 goroutine 可能因为错误导致 panic
				return
			}
			cursor = newCursor

			for i := 0; i < len(res); i += 2 {
				k := res[i]
				v := res[i+1]

				strs := strings.Split(k, ":")
				if len(strs) < 2 {
					continue
				}
				biz := strs[0]
				bizId, _ := strconv.ParseInt(strs[1], 10, 64)
				cnt, _ := strconv.ParseInt(v, 10, 64)

				out <- domain.FavoriteCount{
					Count: cnt,
					Biz:   biz,
					BizId: bizId,
				}
			}

			// 已经扫描完毕
			if cursor == 0 {
				break
			}
		}
	}()

	return out, nil
}

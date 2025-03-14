package dao

import (
	"context"
	"gorm.io/gorm/clause"

	"gorm.io/gorm"

	"github.com/crazyfrankie/favorite/internal/biz/domain"
)

type FavoriteWriteDao struct {
	db *gorm.DB
}

func NewFavoriteWriteDao(db *gorm.DB) *FavoriteWriteDao {
	return &FavoriteWriteDao{db: db}
}

func (d *FavoriteWriteDao) SaveFavoriteCounts(ctx context.Context, counts []domain.FavoriteCount) error {
	cnts := make([]FavoriteCount, len(counts))
	for i, c := range counts {
		cnts[i] = FavoriteCount{
			Biz:   c.Biz,
			BizId: c.BizId,
			Count: c.Count,
		}
	}

	batchSize := 500
	if len(cnts) < batchSize {
		batchSize = len(cnts)
	}

	// 批量写入
	err := d.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "biz"}, {Name: "biz_id"}}, // 唯一键冲突时
			DoUpdates: clause.AssignmentColumns([]string{"count"}),      // 更新 count
		},
	).CreateInBatches(&cnts, batchSize).Error

	return err
}

type FavoriteReadDao struct {
	db *gorm.DB
}

func NewFavoriteReadDao(db *gorm.DB) *FavoriteReadDao {
	return &FavoriteReadDao{db: db}
}

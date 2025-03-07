package dao

import "gorm.io/gorm"

type FavoriteDao struct {
	db *gorm.DB
}

func NewFavoriteDao(db *gorm.DB) *FavoriteDao {
	return &FavoriteDao{db: db}
}

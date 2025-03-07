package dao

import "gorm.io/gorm"

type Favorite struct {
}

type FavoriteDao struct {
	db *gorm.DB
}

func NewFavoriteDao(db *gorm.DB) *FavoriteDao {
	return &FavoriteDao{db: db}
}

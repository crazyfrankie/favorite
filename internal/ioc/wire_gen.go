// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"fmt"
	"github.com/crazyfrankie/favorite/internal/biz/repository"
	"github.com/crazyfrankie/favorite/internal/biz/repository/cache"
	dao2 "github.com/crazyfrankie/favorite/internal/biz/repository/dao"
	"github.com/crazyfrankie/favorite/internal/biz/service"
	"github.com/crazyfrankie/favorite/internal/config"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
)

// Injectors from wire.go:

func InitServer() *service.FavoriteServer {
	cmdable := InitCache()
	favoriteCache := cache.NewFavoriteCache(cmdable)
	db := InitDB()
	favoriteDao := dao2.NewFavoriteDao(db)
	favoriteRepo := repository.NewFavoriteRepo(favoriteCache, favoriteDao)
	favoriteServer := service.NewFavoriteServer(favoriteRepo)
	return favoriteServer
}

// wire.go:

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(config.GetConf().MySQL.DSN, os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DB"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
	})

	db.AutoMigrate(&dao2.FavoriteCount{}, &dao2.UserFavorite{})

	if err != nil {
		panic(err)
	}

	return db
}

func InitCache() redis.Cmdable {
	cli := redis.NewClient(&redis.Options{
		Addr: config.GetConf().Redis.Addr,
	})

	return cli
}

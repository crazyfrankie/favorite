//go:build wireinject

package ioc

import (
	"fmt"
	"os"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/crazyfrankie/favorite/internal/config"
	"github.com/crazyfrankie/favorite/internal/biz/repository"
	"github.com/crazyfrankie/favorite/internal/biz/repository/cache"
	"github.com/crazyfrankie/favorite/internal/biz/repository/dao"
	"github.com/crazyfrankie/favorite/internal/biz/service"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(config.GetConf().MySQL.DSN,
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
	})

	db.AutoMigrate(&dao.FavoriteCount{}, &dao.UserFavorite{})

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

func InitServer() *service.FavoriteServer {
	wire.Build(
		InitDB,
		InitCache,
		dao.NewFavoriteDao,
		cache.NewFavoriteCache,
		repository.NewFavoriteRepo,
		service.NewFavoriteServer,
	)

	return new(service.FavoriteServer)
}

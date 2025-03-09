//go:build wireinject

package ioc

import (
	"fmt"
	"github.com/crazyfrankie/favorite/config"
	"github.com/crazyfrankie/favorite/internal/biz/repository"
	"github.com/crazyfrankie/favorite/internal/biz/repository/cache"
	"github.com/crazyfrankie/favorite/internal/biz/repository/dao"
	"github.com/crazyfrankie/favorite/internal/biz/service"
	"github.com/crazyfrankie/favorite/internal/rpc"
	"os"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

func InitRegistry() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.GetConf().ETCD.EndPoints},
		DialTimeout: time.Second * 2,
	})
	if err != nil {
		panic(err)
	}

	return cli
}

func InitServer() *rpc.Server {
	wire.Build(
		InitDB,
		InitCache,
		InitRegistry,
		dao.NewFavoriteDao,
		cache.NewFavoriteCache,
		repository.NewFavoriteRepo,
		service.NewFavoriteServer,
		rpc.NewServer,
	)

	return new(rpc.Server)
}

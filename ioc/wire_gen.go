// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package ioc

import (
	"fmt"
	"github.com/crazyfrankie/favorite/biz/repository"
	"github.com/crazyfrankie/favorite/biz/repository/cache"
	"github.com/crazyfrankie/favorite/biz/repository/dao"
	"github.com/crazyfrankie/favorite/biz/service"
	"github.com/crazyfrankie/favorite/config"
	"github.com/crazyfrankie/favorite/rpc"
	"github.com/redis/go-redis/v9"
	"go.etcd.io/etcd/client/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"os"
	"time"
)

// Injectors from wire.go:

func InitServer() *rpc.Server {
	cmdable := InitCache()
	favoriteCache := cache.NewFavoriteCache(cmdable)
	db := InitDB()
	favoriteDao := dao.NewFavoriteDao(db)
	favoriteRepo := repository.NewFavoriteRepo(favoriteCache, favoriteDao)
	favoriteServer := service.NewFavoriteServer(favoriteRepo)
	client := InitRegistry()
	server := rpc.NewServer(favoriteServer, client)
	return server
}

// wire.go:

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(config.GetConf().MySQL.DSN, os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DB"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
	})
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

package main

import (
	"context"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/crazyfrankie/favorite/internal/config"
	"github.com/crazyfrankie/favorite/rpc"
)

func main() {
	server := rpc.NewServer(initRegistry())

	g := &run.Group{}

	g.Add(func() error {
		return server.Serve()
	}, func(err error) {
		server.Shutdown()
	})

	favoriteServer := &http.Server{Addr: ":9092"}
	g.Add(func() error {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			rpc.PromRegistry,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		))
		return favoriteServer.ListenAndServe()
	}, func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := favoriteServer.Shutdown(ctx); err != nil {
			log.Printf("failed to shutdown metrics server: %v", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		log.Printf("program interrupted, err:%s", err)
		return
	}
}

func initRegistry() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.GetConf().ETCD.EndPoints},
		DialTimeout: time.Second * 2,
	})
	if err != nil {
		panic(err)
	}

	return cli
}

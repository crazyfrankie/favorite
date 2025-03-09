package main

import (
	"context"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/crazyfrankie/favorite/internal/ioc"
	"github.com/crazyfrankie/favorite/internal/rpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	server := ioc.InitServer()

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

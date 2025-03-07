package rpc

import (
	"net"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/favorite/biz/service"
	"github.com/crazyfrankie/favorite/config"
	"github.com/crazyfrankie/favorite/pkg/registry"
	"github.com/crazyfrankie/favorite/rpc_gen/favorite"
)

type Server struct {
	*grpc.Server
	Port     string
	registry *registry.ServiceRegistry
}

func NewServer(f *service.FavoriteServer, client *clientv3.Client) *Server {
	s := grpc.NewServer()
	favorite.RegisterFavoriteServiceServer(s, f)

	rgy, err := registry.NewServiceRegistry(client)
	if err != nil {
		panic(err)
	}

	return &Server{
		Server:   s,
		Port:     config.GetConf().Server.Port,
		registry: rgy,
	}
}

func (s *Server) Serve() error {
	conn, err := net.Listen("tcp", s.Port)
	if err != nil {
		return err
	}

	err = s.registry.Register()
	if err != nil {
		return err
	}

	return s.Server.Serve(conn)
}

func (s *Server) Shutdown() {
	err := s.registry.UnRegister()
	if err != nil {
		zap.L().Error("Failed to unregister", zap.Error(err))
	}

	s.Server.GracefulStop()
	s.Server.Stop()
}

package service

import (
	"context"

	"github.com/crazyfrankie/favorite/biz/repository"
	"github.com/crazyfrankie/favorite/rpc_gen/favorite"
)

type FavoriteServer struct {
	repo *repository.FavoriteRepo

	favorite.UnimplementedFavoriteServiceServer
}

func NewFavoriteServer(repo *repository.FavoriteRepo) *FavoriteServer {
	return &FavoriteServer{repo: repo}
}

func (f FavoriteServer) FavoriteAction(ctx context.Context, request *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	//TODO implement me
	panic("implement me")
}

package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/crazyfrankie/favorite/internal/biz/repository"
	"github.com/crazyfrankie/favorite/pkg/constants"
	"github.com/crazyfrankie/favorite/rpc_gen/favorite"
)

type FavoriteServer struct {
	repo *repository.FavoriteRepo

	favorite.UnimplementedFavoriteServiceServer
}

func NewFavoriteServer(repo *repository.FavoriteRepo) *FavoriteServer {
	return &FavoriteServer{repo: repo}
}

func (f *FavoriteServer) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (*favorite.FavoriteActionResponse, error) {
	// 校验参数
	action := req.GetActionType()
	if action != constants.FavoriteActionType && action != constants.UnFavoriteActionType {
		return nil, status.Errorf(codes.InvalidArgument, "invalid action type: %d", action)
	}

	// 调用业务方查询内容是否存在（示例代码）
	// video, err := f.videoClient.GetVideoExist(ctx, &video.GetVideoExistRequest{ id: req.GetBizId() })
	// if err != nil { return nil, status.Errorf(codes.NotFound, "video not found") }

	userID, bizID, biz := req.GetUserId(), req.GetBizId(), req.GetBiz()

	if action == constants.FavoriteActionType {
		if err := f.repo.CreateFavorite(ctx, biz, bizID, userID); err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				return nil, status.Errorf(codes.AlreadyExists, "favorite already exists")
			}
			return nil, status.Errorf(codes.Internal, "failed to create favorite: %v", err)
		}
	} else {
		// 直接尝试删除点赞（内部保证幂等）
		if err := f.repo.DeleteFavorite(ctx, biz, bizID, userID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return nil, status.Errorf(codes.NotFound, "favorite not found")
			}
			return nil, status.Errorf(codes.Internal, "failed to delete favorite: %v", err)
		}
	}

	// 成功返回
	return &favorite.FavoriteActionResponse{}, nil
}

// FavoriteList 获取用户的点赞列表
func (f *FavoriteServer) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (*favorite.FavoriteListResponse, error) {
	res, err := f.repo.UserFavoriteElements(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get favorite list: %v", err)
	}

	return &favorite.FavoriteListResponse{BizId: res}, nil
}

// FavoriteCount 获取单个内容的点赞数
func (f *FavoriteServer) FavoriteCount(ctx context.Context, req *favorite.FavoriteCountRequest) (*favorite.FavoriteCountResponse, error) {
	count, err := f.repo.FavoriteCount(ctx, req.GetBiz(), req.GetBizId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get favorite count: %v", err)
	}

	return &favorite.FavoriteCountResponse{Count: count}, nil
}

// IsFavorite 获取用户是否点赞
func (f *FavoriteServer) IsFavorite(ctx context.Context, req *favorite.IsFavoriteRequest) (*favorite.IsFavoriteResponse, error) {
	fav, err := f.repo.IsUserFavorite(ctx, req.GetUserId(), req.GetBizId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get is favorite: %v", err)
	}

	return &favorite.IsFavoriteResponse{Favorite: fav}, nil
}

// UserFavoriteCount 获取用户的点赞总数
func (f *FavoriteServer) UserFavoriteCount(ctx context.Context, req *favorite.UserFavoriteCountRequest) (*favorite.UserFavoriteCountResponse, error) {
	count, err := f.repo.UserFavoriteCount(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user favorite count: %v", err)
	}

	return &favorite.UserFavoriteCountResponse{Count: count}, nil
}

// UserFavoritedCount 获取用户的被点赞总数
func (f *FavoriteServer) UserFavoritedCount(ctx context.Context, req *favorite.UserFavoritedCountRequest) (*favorite.UserFavoritedCountResponse, error) {
	count, err := f.repo.UserFavoritedCount(ctx, req.GetBiz(), req.GetBizId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user favorited count: %v", err)
	}

	return &favorite.UserFavoritedCountResponse{Count: count}, nil
}

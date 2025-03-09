package domain

type UserFavorite struct {
	UserId int64
	Biz    string
	BizId  int64
	Status uint8
}

type FavoriteCount struct {
	UserId int64
	Count  int64
	Biz    string
	BizId  int64
}

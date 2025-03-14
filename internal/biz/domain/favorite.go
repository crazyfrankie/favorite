package domain

type UserFavorite struct {
	UserId int64
	Biz    string
	BizId  int64
	Status uint8
}

type FavoriteCount struct {
	Count int64
	Biz   string
	BizId int64
}

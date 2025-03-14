package dao

type FavoriteCount struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Biz   string `gorm:"uniqueIndex:biz_id_type;type:varchar(128)"`
	BizId int64  `gorm:"uniqueIndex:biz_id_type"`
	Count int64  `gorm:"not null;default:0"`
	Ctime int64  `gorm:"autoCreateTime"` // 时间戳
	Utime int64  `gorm:"autoUpdateTime"`
}

type UserFavorite struct {
	Id     int64  `gorm:"primaryKey,autoIncrement"`
	UserId int64  `gorm:"index:idx_uid"` // 用户 ID
	Biz    string `gorm:"index:idx_biz;type:varchar(128)"`
	BizId  int64  `gorm:"index:idx_biz"`
	Status uint8  `gorm:"not null;default:1"` // 0: 取消点赞, 1: 点赞
	Ctime  int64  `gorm:"autoCreateTime"`
	Utime  int64  `gorm:"autoUpdateTime"`
}

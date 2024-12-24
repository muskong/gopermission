package entity

import "time"

type Blacklist struct {
	ID         int    `gorm:"primarykey"`
	Name       string `gorm:"type:varchar(50);index"`
	Phone      string `gorm:"type:varchar(20);index"`
	IDCard     string `gorm:"type:varchar(18);index"`
	Email      string `gorm:"type:varchar(100)"`
	Address    string `gorm:"type:varchar(255)"`
	Remark     string `gorm:"type:text"`
	Status     int    `gorm:"type:int;default:0"` // 0:待审核 1:已通过 2:已拒绝
	MerchantID int    `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time `gorm:"index"`
}
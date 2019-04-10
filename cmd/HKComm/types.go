package HKComm

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	UserName string `gorm:"type:varchar(255);not null;unique_index" form:"username" validate:"alphanumunicode;max=255"`
	PassWord string `gorm:"type:varchar(255);not null" form:"password" validate:"alphanumunicode;min=8"`
	Email string `validate:"email"`
}

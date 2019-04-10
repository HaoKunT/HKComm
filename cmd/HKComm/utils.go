package HKComm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/sessions"
)


/*
	状态码:
		`0`: ok
		`-1`: 用户不存在或密码错误
		``
 */

const (
	OK = -iota
	UserNotFoundOrPasswordError
)

var Msg = map[int]string{
	OK: "ok",
	UserNotFoundOrPasswordError: "user not found or wrong password!",
}

var db *gorm.DB
var sess *sessions.Sessions

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

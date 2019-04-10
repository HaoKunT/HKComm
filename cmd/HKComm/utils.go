package HKComm

import "fmt"


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



func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

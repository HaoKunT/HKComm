package hkcomm

import (
	"github.com/kataras/iris"
)

/*
实现好友相关的接口
好友部分的逻辑
首先需要搜索好友，然后根据搜索到的好友发送好友添加请求，等待对方认证后完成添加好友的过程，并向双方发送`你已经添加对方为好友`的提示
*/

/*
 发送添加好友申请的接口
*/
//func AddFriend(ctx iris.Context) {
//	friendID, err := ctx.Params().GetUint("friendid")
//	if checkError(err) != nil {
//		ctx.JSON(returnStruct{
//			Status:  iris.StatusOK,
//			Code:    ParamsError,
//			Message: Msg[ParamsError],
//			Error:   err.Error(),
//		})
//		return
//	}
//
//}

// 通过ID搜索用户
func SearchID(ctx iris.Context) {
	userid, err := ctx.Params().GetUint("userid")
	if checkError(err) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ParamsError,
			Message: Msg[ParamsError],
			Error:   errorString(err.Error()),
		})
		return
	}
	var users []User
	if err = checkError(db.Where("id = ?", userid).Select("username, id, email").Find(&users).Error); err != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    NotFound,
			Message: Msg[NotFound],
			Error:   errorString(err.Error()),
		})
		return
	}
	ctx.JSON(returnStruct{
		Status:  iris.StatusOK,
		Code:    OK,
		Message: Msg[OK],
		Data:    users,
	})
	return
}

//通过名字搜索用户
func SearchName(ctx iris.Context) {
	username := ctx.Params().GetString("username")
	var users []User
	if err := checkError(db.Where("username = ?", username).Select("username, id, email").Find(&users).Error); err != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    NotFound,
			Message: Msg[NotFound],
			Error:   errorString(err.Error()),
		})
		return
	}
	ctx.JSON(returnStruct{
		Status:  iris.StatusOK,
		Code:    OK,
		Message: Msg[OK],
		Data:    users,
	})
	return
}

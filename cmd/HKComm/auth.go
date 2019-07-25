package hkcomm

import (
	"github.com/kataras/iris"
	"gopkg.in/go-playground/validator.v9"
)

/*
使用邮箱和密码进行登录
*/
func login(ctx iris.Context) {
	s := sess.Start(ctx)
	formUser := User{}
	if checkError(ctx.ReadForm(&formUser)) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
	vali := validator.New()
	if checkError(vali.Struct(&formUser)) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
	sqlUser := User{
		PassWord: "",
	}
	if checkError(db.Where("email = ?", formUser.Email).First(&sqlUser).Error) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
	if sqlUser.PassWord == "" {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    UserNotFoundOrPasswordError,
			Message: Msg[UserNotFoundOrPasswordError],
		})
		return
	} else if sqlUser.PassWord != formUser.PassWord {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    UserNotFoundOrPasswordError,
			Message: Msg[UserNotFoundOrPasswordError],
		})
		return
	}
	s.Set("userid", int(sqlUser.ID))
	s.Set("authenticated", true)
	ctx.JSON(returnStruct{
		Status:  iris.StatusOK,
		Code:    OK,
		Message: Msg[OK],
	})
}

// 登出的时候将session内的值清空
func logout(ctx iris.Context) {
	s := sess.Start(ctx)
	s.Clear()
	ctx.JSON(returnStruct{
		Status:  iris.StatusOK,
		Code:    OK,
		Message: Msg[OK],
	})
}

func secret(ctx iris.Context) {
	// Print secret message
	ctx.WriteString("The cake is a lie!")
}

func register(ctx iris.Context) {
	/*
		registe the user
	*/
	formUser := User{}
	if checkError(ctx.ReadForm(&formUser)) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
	vali := validator.New()
	if checkError(vali.Struct(&formUser)) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
	if checkError(db.Create(&formUser).Error) != nil {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    ServerError,
			Message: Msg[ServerError],
		})
		return
	}
}

func auth(ctx iris.Context) {
	if auth, _ := sess.Start(ctx).GetBoolean("authenticated"); !auth {
		ctx.JSON(returnStruct{
			Status:  iris.StatusOK,
			Code:    Unauthoried,
			Message: Msg[Unauthoried],
		})
		ctx.StatusCode(iris.StatusForbidden)
		return
	}
	ctx.Next()
}

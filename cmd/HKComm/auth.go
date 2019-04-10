package HKComm

import (
	"github.com/kataras/iris"
	"gopkg.in/go-playground/validator.v9"
)



func login(ctx iris.Context) {
	s := sess.Start(ctx)
	formUser := User{}
	checkError(ctx.ReadForm(&formUser))
	vali := validator.New()
	if err := vali.Struct(&formUser); err != nil {
		panic(err)
	}
	sqlUser := User{
		PassWord: "",
	}
	checkError(db.Where("username = ?", formUser.UserName).Find(&sqlUser).Error)
	if sqlUser.PassWord == "" {
		ctx.JSON(iris.Map{
			"status":  iris.StatusOK,
			"code":    UserNotFoundOrPasswordError,
			"message": Msg[UserNotFoundOrPasswordError],
		})
		return
	} else if sqlUser.PassWord == formUser.PassWord {
		ctx.JSON(iris.Map{
			"status":  iris.StatusOK,
			"code":    OK,
			"message": Msg[OK],
		})
		return
	}
	s.Set("userid", sqlUser.ID)
	s.Set("authenticated", true)
}

func logout(ctx iris.Context) {
	s := sess.Start(ctx)
	s.Set("authenticated", false)
	ctx.JSON(iris.Map{
		"status":  iris.StatusOK,
		"code":    OK,
		"message": Msg[OK],
	})
	ctx.StatusCode(iris.StatusOK)
}

func secret(ctx iris.Context) {

	// Check if user is authenticated
	if auth, _ := sess.Start(ctx).GetBoolean("authenticated"); !auth {
		ctx.StatusCode(iris.StatusForbidden)
		return
	}

	// Print secret message
	ctx.WriteString("The cake is a lie!")
}

func register(ctx iris.Context) {
	/*
	registe the user
	 */
	formUser := User{}
	checkError(ctx.ReadForm(&formUser))
	vali := validator.New()
	if err := vali.Struct(&formUser); err != nil {
		panic(err)
	}
	db.Create(&formUser)
}
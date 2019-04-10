package HKComm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/go-playground/validator.v9"
)

func CreateSuperUser() (err error) {
	user := User{}
	fmt.Printf("Please enter username (default: HKComm):")
	if _, err = fmt.Scanln(&user.UserName); err != nil {
		return
	}
	var tmpPassword1 string
	fmt.Printf("Please enter password:")
	if _, err = fmt.Scanln(&tmpPassword1); err != nil {
		return
	}
	var tmpPassword2 string
	fmt.Printf("Please enter password again:")
	if _, err = fmt.Scanln(&tmpPassword2); err != nil {
		return
	}
	if tmpPassword1 != tmpPassword2 {
		return fmt.Errorf("create super user: the password is not same")
	}
	user.PassWord = tmpPassword1
	fmt.Printf("Please enter email:")
	if _, err = fmt.Scanln(&user.Email); err != nil {
		return
	}
	vali := validator.New()
	if err = vali.Struct(&user); err != nil {
		return
	}
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	defer func() {
		if err = db.Close(); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		return err
	}
	return db.Create(&user).Error
}

func InitDatabase() (err error) {
	db, err := gorm.Open("sqlite3", "db.sqlite3")
	defer func() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}()
	if err != nil {
		return
	}
	if db.HasTable(&User{}) {
		fmt.Println("Table user is exists!")}
	db.AutoMigrate(&User{})
	fmt.Println("Init database sucessful")
	return nil
}

func Server()  {
	app := iris.Default()

	sess := sessions.New(sessions.Config{
		Cookie:       "HKCommSession",
		Expires:      -1,
		AllowReclaim: true,
	})

	db, err := gorm.Open("sqlite3", "db.sqlite3")
	checkError(err)
	defer db.Close()
	db.SingularTable(true)

	app.Post("/login", func(ctx iris.Context) {
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
	})

	app.Post("/logout", func(ctx iris.Context) {
		s := sess.Start(ctx)
		s.Clear()
		ctx.JSON(iris.Map{
			"status":  iris.StatusOK,
			"code":    OK,
			"message": Msg[OK],
		})
		ctx.StatusCode(iris.StatusOK)
	})

	app.Get("/secret", func(ctx iris.Context) {

		// Check if user is authenticated
		if auth, _ := sess.Start(ctx).GetBoolean("authenticated"); !auth {
			ctx.StatusCode(iris.StatusForbidden)
			return
		}

		// Print secret message
		ctx.WriteString("The cake is a lie!")
	})

	app.Get("/ping", func(ctx iris.Context) {
		ctx.WriteString("Welcome!")
	})

	app.Run(iris.Addr(":8080"))
}

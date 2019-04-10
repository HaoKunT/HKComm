package HKComm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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
	user.Email = strings.Trim(user.Email, "\r\n")
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

	var err error

	sess = sessions.New(sessions.Config{
		Cookie:       "HKCommSession",
		Expires:      -1,
		AllowReclaim: true,
	})

	db, err = gorm.Open("sqlite3", "db.sqlite3")
	checkError(err)
	defer db.Close()
	db.SingularTable(true)

	app.Post("/login", login)

	app.Post("/logout", logout)

	app.Get("/secret", secret)

	app.Get("/ping", func(ctx iris.Context) {
		ctx.WriteString("pong")
	})

	app.Post("/register", register)

	app.Run(iris.Addr(":8080"))
}

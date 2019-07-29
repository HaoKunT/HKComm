package hkcomm

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"gopkg.in/go-playground/validator.v9"
)

func CreateSuperUser() (err error) {
	defer func() {
		SafeExit()
	}()
	user := User{}
	fmt.Printf("Please enter username (default: HKComm):")
	inputReader := bufio.NewReader(os.Stdin)
	var str string
	if str, err = inputReader.ReadString('\n'); err != nil {
		return
	}
	str = strings.Trim(str, "\r\n")
	if str == "" {
		str = "HKComm"
	}
	user.UserName = str
	var tmpPassword1 []byte
	fmt.Printf("Please enter password:")
	if tmpPassword1, err = gopass.GetPasswd(); err != nil {
		return
	}
	var tmpPassword2 []byte
	fmt.Printf("Please enter password again:")
	if tmpPassword2, err = gopass.GetPasswd(); err != nil {
		return
	}
	s_tmpPassword1 := string(tmpPassword1)
	s_tmpPassword2 := string(tmpPassword2)
	if s_tmpPassword1 != s_tmpPassword2 {
		return fmt.Errorf("create super user: the password is not same")
	}
	user.PassWord = s_tmpPassword1
	fmt.Printf("Please enter email:")
	if user.Email, err = inputReader.ReadString('\n'); err != nil {
		return
	}
	user.Email = strings.Trim(user.Email, "\r\n")
	vali := validator.New()
	if err = vali.Struct(&user); err != nil {
		return
	}
	return db.Create(&user).Error
}

func InitDatabase() (err error) {
	defer func() {
		SafeExit()
	}()
	db.Debug().AutoMigrate(&User{}, &Group{}, &File{}, &communicationData{})
	fmt.Println("Init database sucessful")
	return nil
}

func Server() {
	loadSDB()
	loadDB()
	defer func() {
		SafeExit()
	}()
	app := iris.Default()

	sess = sessions.New(sessions.Config{
		Cookie:       "HKCommSession",
		Expires:      -1,
		AllowReclaim: true,
	})

	db.SingularTable(true)

	setUpWebsocket(app)

	msgCh()

	needAuth := app.Party("/", auth)

	app.Post("/login", login)

	app.Post("/logout", logout)

	needAuth.Get("/secret", secret)

	app.Get("/ping", func(ctx iris.Context) {
		ctx.WriteString("pong")
	})

	app.Post("/register", register)

	needAuth.Get("/search/id/{userid:uint}", SearchID)

	needAuth.Get("/search/username/{username:string max(255)}", SearchName)

	//app.Post("/addfriend/{friendid:uint}", auth, AddFriend)

	needAuth.Get("/init/history-cds", getNew100)

	needAuth.Post("/upload", uploadFile)

	app.Run(iris.Addr(":8080"))
}

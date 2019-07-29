package hkcomm

import (
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/sessions"
	"github.com/kataras/iris/sessions/sessiondb/badger"
	"github.com/streadway/amqp"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGINT)
	go func() {
		<- sigs
		SafeExit()
	}()
	defer func() {
		if err := recover(); err != nil {
			SafeExit()
			panic(err)
		}
	}()
	var err error
	// 加载配置文件
	load()
	sess = sessions.New(sessions.Config{
		Cookie:       "HKCommSessions",
		Expires:      -1,
		AllowReclaim: true,
	})
	sess.UseDatabase(sdb)

	connAMQP, err = amqp.Dial(rabbitmqURL)
	if checkError(err) != nil {
		log.Fatalf("Fail to connect RabbitMQ: %s", err)
	}
	chAMQP, err = connAMQP.Channel()
	if checkError(err) != nil {
		log.Fatalf("Fail to create channel: %s", err)
	}
	// 声明exchange
	err = chAMQP.ExchangeDeclare(
		"hkcomm",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if checkError(err) != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
	}
	// 需要Confirm
	confirms = chAMQP.NotifyPublish(make(chan amqp.Confirmation, 100))
	if err = checkError(chAMQP.Confirm(false)); err != nil {
		log.Fatalf("confirm.select: %s", err)
	}
	syncTag = 0
	Confirm = make(chan *unConfirmTag, 100)
	// Qos
	err = chAMQP.Qos(10, 0, false)
	if checkError(err) != nil {
		log.Fatalf("Failed to Qos: %s", err)
	}
	msgInCh = make(chan communicationData, msgInBuffer)
	msgOutCh = make(chan communicationData, msgOutBuffer)
	// 检验目录是否存在
	if !Exists(filePath) || !IsDir(filePath) {
		if err := checkError(os.Mkdir(filePath, 0755)); err != nil {
			log.Fatalf("make filepath: %s, %s", filePath, err)
		}
	}

}

func loadSDB()  {
	var err error
	sdb, err = badger.New(sdbPath)
	if err != nil {
		log.Fatalf("session database error on path %s, %s", sdbPath, err)
	}
}

func loadDB()  {
	var err error
	db, err = gorm.Open("sqlite3", "db.sqlite3")
	if checkError(err) != nil {
		log.Fatalf("Open database error: %s", err)
	}
	db.SingularTable(true)
}
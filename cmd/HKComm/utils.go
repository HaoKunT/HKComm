package hkcomm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
	"github.com/kataras/iris/sessions/sessiondb/badger"
	"github.com/streadway/amqp"
	"log"
	"sync"
)

/*
	状态码:
		`0`: ok
		`-1`: 用户不存在或密码错误
		``
*/
const (
	// OK :正常
	OK = -iota
	// UserNotFoundOrPasswordError :没找到用户或者密码错误
	UserNotFoundOrPasswordError
	// Unauthoried :没有认证 403
	Unauthoried
	// ServerError :服务器错误 500
	ServerError
	// NotFound :404
	NotFound
	// ParamsError :参数错误
	ParamsError
	// FileError :文件的md5检查错误
	FileError
)

// Msg :错误所对应的文字信息
var Msg = map[int]string{
	OK:                          "ok",
	UserNotFoundOrPasswordError: "user not found or wrong password!",
	Unauthoried:                 "unauthoried",
	ServerError:                 "server error",
	NotFound:                    "not found",
	ParamsError:                 "params error",
	FileError:                   "file error",
}

// debug
var Isdebug bool

// database
var db *gorm.DB
// session database
var sdb *badger.Database
var sdbPath string

// session对象
var sess *sessions.Sessions

// 维持所有的websocket连接
type connPool struct {
	sync.RWMutex
	conn map[uint][]*connection
}

// 单例模式
var conns *connPool
var connOnce sync.Once

// GetConns :获取连接池
func GetConns() *connPool {
	connOnce.Do(func() {
		var lock sync.RWMutex
		conn := make(map[uint][]*connection)
		conns = &connPool{
			lock,
			conn,
		}
	})
	return conns
}

// RabbitMQ的URL
var rabbitmqURL string
// connAMQP 连接RabbitMQ的连接
var connAMQP *amqp.Connection

// channelAMQP 连接RabbitMQ的通道
var chAMQP *amqp.Channel
// confirms
var confirms <- chan amqp.Confirmation
// channel confirm tag sync
var syncTag uint64
// offlineConfirm
var Confirm chan *unConfirmTag

// 处理通道
var msgInBuffer uint
var msgInCh chan communicationData

// 发送通道
var msgOutBuffer uint
var msgOutCh chan communicationData

// 处理线程数量
var msgNumber uint

// 存储文件所在的目录
var filePath string
// 文件元数据缓存
// 文件元数据缓存结构
type fileCache struct{
	sync.Mutex
	cache map[string]*communicationData
}
// 单例模式
var fileOnce sync.Once
var filecache *fileCache

// GetFilecache :获取文件缓存
func GetFilecache() *fileCache {
	fileOnce.Do(func() {
		var lock sync.Mutex
		cache := make(map[string]*communicationData)
		filecache = &fileCache{
			lock,
			cache,
		}
	})
	return filecache
}

// check the error
func checkError(err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

// SafeExit : SafeExit the server
var safeExit sync.Once
func SafeExit()  {
	safeExit.Do(func() {
		sdb.Close()
		connAMQP.Close()
		db.Close()
		log.Println("安全退出")
	})
}

// get userid by session
func getUIDByContext(ctx iris.Context) (int, error) {
	id, err := sess.Start(ctx).GetInt("userid")
	if checkError(err) != nil {
		return -1, err
	} else {
		return id, nil
	}
}

// send message
func sendMsg(id uint, cd *communicationData)  {
	conns := GetConns()
	conns.RLock()
	for _, c := range conns.conn[cd.To] {
		i := 0
		var err error
	sendretry: {
		if i < 5 {
			if err = c.Emit("msg", *cd); err != nil {
				checkError(fmt.Errorf("send message error: %s", err))
				i += 1
				goto sendretry
			}
		} else {
			log.Printf("send to user %d error: %s", id, err)
		}
	}
	}
	conns.RUnlock()
}

// get user by session
func getUserByContext(ctx iris.Context) (user *User, err error) {
	id ,err := getUIDByContext(ctx)
	if err != nil {
		return nil, err
	}
	user = &User{}
	user.ID = uint(id)
	user.GetInfo()
	return user, nil
}

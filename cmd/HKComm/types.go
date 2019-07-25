package hkcomm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kataras/iris/websocket"
)

// wrapper the websocket.Connection
type connection struct {
	websocket.Connection
	senderID uint
}


type StringBaseModel struct {
	ID        string `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type MyBaseModel struct {
	ID        uint `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// User struct
type User struct {
	MyBaseModel
	UserName     string   `json:"username" gorm:"column:username;type:varchar(255);index" form:"username" validate:"omitempty,alphanumunicode,max=255,required"`
	PassWord     string   `json:"omitempty,password" gorm:"column:password;type:varchar(255);not null" form:"password" validate:"alphanumunicode|containsany=!@#?$%^&,min=8"`
	Email        string   `json:"email" gorm:"type:varchar(255);not null;unique_index" validate:"omitempty,email" form:"email"`
	Friends      []*User  `json:"omitempty,friends" gorm:"many2many:friendships;association_jointable_foreignkey:friend_id"`
	Groups       []*Group `json:"omitempty,groups" gorm:"many2many:user_groups;"`
	SendFiles    []*File  `json:"omitempty,send_files"`
	ReceiveFiles []*File  `json:"omitempty,receive_files"`
}

// Group struct
type Group struct {
	MyBaseModel
	Name  string  `gorm:"type:varchar(255);not null;index" form:"username" validate:"alphanumunicode,max=255,required"`
	Users []*User `gorm:"many2many:user_groups;"`
	Files []*File
}

// Files struct
type File struct {
	MyBaseModel
	Name     string `gorm:"type:varchar(255);not null;index" form:"filename" validate:"max=255,required" json:"name"`
	Md5      string `gorm:"type:varchar(64);not null;index"`
	Sender   uint
	Receiver uint
	Isgroup  bool `json:"isgroup"`
	CDataID string
}

// communicationData
// if the communication is transform the file, the Content is the info of the file
// the communication data should be save into the database in the future
type communicationData struct {
	// 数据在哪个连接上
	connection *connection `json:"-" gorm:"-"`
	StringBaseModel
	From uint `json:"from" gorm:"from"`
	To   uint `json:"to" gorm:"to"`
	// 发送对象是否是一个群
	Isgroup bool   `json:"isgroup" gorm:"isgroup"`
	Message string `json:"content" gorm:"message"`
	// 是否为文件
	Isfile bool `json:"isfile" gorm:"isfile"`
	// 文件类型（文件信息由这个传输，文件本身不由此传输）
	File File `json:"file,omitempty" gorm:"foreignkey:CDataID"`
	// 时间戳
	Timestamp int64 `json:"timestamp" gorm:"timestamp"`
	// 同步位
	Sync uint64`json:"sync" gorm:"-"`
}

// 返回值
type returnStruct struct {
	Status  uint        `json:"status"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   errorString `json:"error,omitempty"`
}

type errorString = string

//func (es errorString)MarshalJSON() ([]byte, error) {
//	// 在不处于debug模式下，将Error字段置空
//	if !Isdebug {
//		es = ""
//	}
//	var buf bytes.Buffer
//	buf.Grow(2+len(es))
//	buf.WriteString("\"")
//	buf.WriteString(string(es))
//	buf.WriteString("\"")
//	return buf.Bytes(), nil
//}

// 确认消息是否收到的回执
type receipt struct {
	ID string `json:"id"`
	Sync uint64 `json:"sync"`
	Status  uint `json:"status"`
	Message string `json:"message"`
}

// 用于存储尚未confirm的结构
type unConfirmTag struct{
	sync uint64
	sender *connection
	confirmTag uint64
	data *communicationData
}

// GetGInfo 根据用户id查询数据(包含群组信息)
func (user *User) GetGInfo() error {
	if err := db.Where("id = ?", user.ID).Preload("Groups").Find(user).Error; err != nil {
		err = fmt.Errorf("Wrong userid: %d, error info: %s", user.ID, err)
		checkError(err)
		return err
	}
	return nil
}

// GetInfo 根据用户id查询数据(不包含群组信息)
func (user *User) GetInfo() error {
	if err := db.Where("id = ?", user.ID).Find(user).Error; err != nil {
		err = fmt.Errorf("Wrong userid: %d, error info: %s", user.ID, err)
		checkError(err)
		return err
	}
	return nil
}

// GetUInfo 根据群组id查询数据(包含用户信息)
func (group *Group) GetUInfo() error {
	if err := db.Where("id = ?", group.ID).Preload("Users").Find(group).Error; err != nil {
		err = fmt.Errorf("Wrong group id: %d, error info: %s", group.ID, err)
		checkError(err)
		return err
	}
	return nil
}

// GetInfo 根据群组id查询数据(不包含用户信息)
func (group *Group) GetInfo() error {
	if err := db.Where("id = ?", group.ID).Find(group).Error; err != nil {
		err = fmt.Errorf("Wrong group id: %d, error info: %s", group.ID, err)
		checkError(err)
		return err
	}
	return nil
}

// Exist 判断是否存在这个用户
func (user *User) Exist() bool {
	if user.GetInfo() != nil {
		return true
	} else {
		return false
	}
}

// 根据收到的消息返回发送者相关信息
// 参数detail: false时代表只查询用户，不查询群组消息
func (cd *communicationData) GetSender(detail bool) (*User, error) {
	var user User
	user.ID = cd.From
	var err error
	if detail {
		err = user.GetGInfo()
	} else {
		err = user.GetInfo()
	}
	if checkError(err) != nil {
		return nil, fmt.Errorf("No sender user %d, %s", cd.From, err)
	}
	return &user, nil
}

// 根据收到的消息返回接收者的相关信息
// 参数detail: false时代表只查询用户，不查询群组消息
func (cd *communicationData) GetReceiver(detail bool) (*User, error) {
	var user User
	user.ID = cd.To
	var err error
	if detail {
		err = user.GetGInfo()
	} else {
		err = user.GetInfo()
	}
	if checkError(err) != nil {
		return nil, fmt.Errorf("No receiver user %d, %s", cd.From, err)
	}
	return &user, nil
}


func (cd *communicationData)generateID() {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d %d %v %s", cd.From, cd.To, cd.Timestamp, cd.Message)))
	cd.ID = hex.EncodeToString(h.Sum(nil))
}

// SaveMsg
// 用于保存消息至数据库
func (cd *communicationData)SaveMsg() error {
	if err := checkError(db.Create(cd).Error); err != nil {
		return err
	}
	return nil
}

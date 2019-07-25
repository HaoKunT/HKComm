package hkcomm

import (
	"log"
	"os"
	"runtime"

	"github.com/Unknwon/goconfig"
)

func load() {
	cfg, err := goconfig.LoadConfigFile("config.conf")
	if checkError(err) != nil {
		log.Fatalf("Loading config file error: %s", err)
	}
	// 读取RabbitMQ所在位置
	rabbitmqURL, err = cfg.GetValue("HKChannel", "URL")
	if checkError(err) != nil {
		log.Fatalf("Read RabbitMQ URL error: %s", err)
	}
	// 读取处理通道缓存大小
	msgInBufferLoad, err := cfg.Int("HKChannel", "msgInBuffer")
	if checkError(err) != nil || msgInBufferLoad <= 0 {
		msgInBuffer = 1000
	} else {
		msgInBuffer = uint(msgInBufferLoad)
	}
	// 读取发送通道缓存大小
	msgOutBufferLoad, err := cfg.Int("HKChannel", "msgOutBuffer")
	if checkError(err) != nil || msgOutBufferLoad <= 0 {
		msgOutBuffer = 1000
	} else {
		msgOutBuffer = uint(msgOutBufferLoad)
	}
	// 读取处理线程的数量
	msgNumberLoad, err := cfg.Int("HKChannel", "msgNumber")
	if checkError(err) != nil || msgNumberLoad <= 0 {
		msgNumber = uint(runtime.NumCPU())
	} else {
		msgNumber = uint(msgNumberLoad)
	}
	// 读取存储文件所在的位置
	filePathLoad, err := cfg.GetValue("file", "path")
	if checkError(err) != nil {
		filePath = "./file"
	} else {
		filePath = filePathLoad
	}
	// 读取session数据库所在的位置
	sdbPathLoad, err := cfg.GetValue("db", "sdbPath")
	if checkError(err) != nil {
		sdbPath = "./sdb"
	} else {
		sdbPath = sdbPathLoad
	}
	// 是否开启debug模式(debug模式将会在发生错误时返回详细错误信息)
	// 默认不开启
	debugLoad, err := cfg.Bool("global", "debug")
	if checkError(err) == nil {
		Isdebug = debugLoad
	}
}

// Exists: 判断所给路径的目录或文件是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// IsDir: 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile: 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

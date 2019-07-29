/**
* @Author: HaoKunT
* @Date: 2019/7/24 0:27
* @File: views.go
*/
package hkcomm

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/kataras/iris"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
)

// 获取最新的100条消息
func getNew100(ctx iris.Context)  {
	id ,err := getUIDByContext(ctx)
	if err != nil {
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: ServerError,
			Message: Msg[ServerError],
			Error: errorString(err.Error()),
		})
		return
	}
	var cds []communicationData
	if err := checkError(db.Where("\"to\" = ?", id).Order("created_at desc").Limit(100).Find(&cds).Error); err != nil {
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: ServerError,
			Message: Msg[ServerError],
			Error: errorString(err.Error()),
		})
		return
	}
	ctx.JSON(returnStruct{
		Status: iris.StatusOK,
		Code: OK,
		Message: Msg[OK],
		Data: cds,
	})
}

// 上传文件的接口
func uploadFile(ctx iris.Context)  {
	user, err := getUserByContext(ctx)
	if err != nil {
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: ServerError,
			Message: Msg[ServerError],
			Error: err.Error(),
		})
		return
	}
	fileId := ctx.FormValue("id")
	filename := ctx.FormValue("filename")
	filecache = GetFilecache()
	filecache.Lock()
	msg, ok := filecache.cache[fileId]
	delete(filecache.cache, fileId)
	filecache.Unlock()
	// 检查是否存在这个文件消息
	if !ok {
		err := fmt.Errorf("no fileId: %s", fileId)
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: NotFound,
			Message: Msg[NotFound],
			Error: err.Error(),
		})
		return
	}
	// 检查文件名是否相符
	if filename != msg.File.Name {
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: FileError,
			Message: Msg[FileError],
			Error: fmt.Sprintf("filename = %s, msg.File.Name = %s", filename, msg.File.Name),
		})
		return
	}
	// 检查用户是否相符
	if user.ID != msg.From {
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: FileError,
			Message: Msg[FileError],
			Error: fmt.Sprintf("user.ID = %d, msg.from = %d", user.ID, msg.From),
		})
		return
	}
	h := md5.New()
	var allerror error
	allerror = nil
	if _, err := ctx.UploadFormFiles("./file", func(ctx iris.Context, header *multipart.FileHeader) {
		fullpath := filepath.Join("./file", header.Filename)
		if err := os.MkdirAll(path.Dir(fullpath), 0666); err != nil {
			allerror = fmt.Errorf("%s\n%s", allerror, err)
			return
		}
		src, err := header.Open()
		if err != nil {
			allerror = fmt.Errorf("%s\n%s", allerror, err)
			return
		}
		defer src.Close()
		buf := bufio.NewReader(src)
		var fb []byte
		_, err = buf.Read(fb)
		if err != nil {
			allerror = fmt.Errorf("%s\n%s", allerror, err)
			return
		}
		h.Write(fb)
	}); err != nil {
		allerror = fmt.Errorf("%s\n%s", allerror, err)
	}
	defer ctx.Request().MultipartForm.RemoveAll()
	// 检查存储过程中有没有错误
	if checkError(allerror) != nil {
		allerror = fmt.Errorf("%s\n%s",
			allerror,
			os.RemoveAll(filepath.Join("./file", filename)),
		)
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: ServerError,
			Message: Msg[ServerError],
			Error: allerror.Error(),
		})
		return
	}
	// 检查md5是否对得上
	md5v := hex.EncodeToString(h.Sum(nil))
	if md5v != msg.File.Md5 {
		err := fmt.Errorf("msg.md5 = %s, calculate md5 = %s", msg.File.Md5, md5v)
		checkError(err)
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: FileError,
			Message: Msg[FileError],
			Error: err.Error(),
		})
		os.RemoveAll(filepath.Join("./file", filename))
		return
	}
	// 为二级目录的文件（目录）重命名
	if err := os.Rename(filepath.Join("./file", filename), filepath.Join("./file", md5v)); err != nil {
		checkError(err)
		ctx.JSON(returnStruct{
			Status: iris.StatusOK,
			Code: FileError,
			Message: Msg[FileError],
			Error: err.Error(),
		})
		os.RemoveAll(filepath.Join("./file", filename))
		return
	}
	// 检查完毕，此时将相应的msg推送至消息队列
	msgOutCh <- *msg
	ctx.JSON(returnStruct{
		Status: iris.StatusOK,
		Code: OK,
		Message: Msg[OK],
	})
}

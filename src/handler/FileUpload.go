package handler

import (
	"fmt"
	"io"
	"net/http"
	"service"
	"utils"
)

func FileUploadHex(w http.ResponseWriter, r *http.Request) {
	utils.Debugf("fileUpload---", *r)
	//strAction := r.FormValue("action")
	uid := GetUID(r)

	fileService := service.GetFileService()
	strFileName := r.FormValue("fileName")
	utils.Debugf("strFileName= %s", strFileName)

	//删掉旧文件
	deleteErr := fileService.DeleteFiles(fileService.GetUserUploadFileDirTemp(uid) + strFileName)
	if deleteErr != nil{
		utils.Debugf("fileUpload删除文件失败 %s", deleteErr)
		return
	}
	filepath := fileService.GetUserUploadFileDirTemp(uid) + strFileName
	f, err := fileService.GetFile(filepath)
	if err != nil {
		defer func(){
			if f != nil {
				f.Close()
			}
		}()
		f, err = fileService.CreateFile(fileService.GetUserUploadFileDirTemp(uid), strFileName)
	}
	if err != nil {//打开或创建失败
		utils.Debugf("fileUpload打开或创建失败 %s", filepath)
		return
	}
	encoding := r.FormValue("encoding")
	utils.Debugf("encoding= %s", encoding)
	strFileContent := r.FormValue("fileContent")
	//utils.Debugf("strFileContent= %s", strFileContent)
	utf8Bytes, _ := utils.GetFromHex(strFileContent)
	if utf8Bytes == nil {
		utils.Debugf("utf8Bytes is nil")
		return
	}
	if encoding=="" || encoding=="utf-8" || encoding=="UTF-8"{//默认utf8
		strFileContent = utils.GetFromUtf8(utf8Bytes)
		utils.Debugf("utf8= %s", strFileContent)
	}else{//否则二进制 encoding=="binary"
		strFileContent = string(utf8Bytes)
	}
	//.Copy(f, file)
	io.WriteString(f, strFileContent)

	retMap := GetRetMap(uid)
	retMap["filepath"] = filepath
	retMap["result"] = "0" //0成功

	fmt.Fprint(w, RetMap2String(retMap))
}

func FileUploadStream(w http.ResponseWriter, r *http.Request) {
	utils.Debugf("fileUpload---", *r)
	//strAction := r.FormValue("action")
	uid := GetUID(r)

	file, fileHead, err := r.FormFile("file0")

	fileService := service.GetFileService()
	strFileName := fileHead.Filename //r.FormValue("fileName")
	utils.Debugf("strFileName= %s", strFileName)

	//删掉旧文件
	deleteErr := fileService.DeleteFiles(fileService.GetUserUploadFileDirTemp(uid) + strFileName)
	if deleteErr != nil{
		utils.Debugf("fileUpload删除文件失败 %s", deleteErr)
		return
	}
	filepath := fileService.GetUserUploadFileDirTemp(uid) + strFileName
	f, err := fileService.GetFile(filepath)
	if err != nil {
		defer func(){
			if f != nil {
				f.Close()
			}
		}()
		f, err = fileService.CreateFile(fileService.GetUserUploadFileDirTemp(uid), strFileName)
	}
	if err != nil {//打开或创建失败
		utils.Debugf("fileUpload打开或创建失败 %s", filepath)
		return
	}
	io.Copy(f, file)

	retMap := GetRetMap(uid)

	fmt.Fprint(w, RetMap2String(retMap))
}


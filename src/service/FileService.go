package service

import (
	//"time"
	//"strconv"
	"os"
	"utils"
)

type FileService struct {
	appRoot string
}

func (this *FileService) SetApplicationRootDir(workDir string) {
	this.appRoot = workDir
}

func (this *FileService) GetFile(fullpath string) (*os.File, error) {
	return os.Open(fullpath)
}
func (this *FileService) CreateFile(dir string, strFileName string) (*os.File, error) {
	//todo Ray 检查是否存在
	//f, err := this.GetFile(dir + strFileName)
	//if err != nil {
	//
	//}
	utils.Debugf("CreateFile---", dir + strFileName)
	this.DeleteFiles(dir + strFileName)//默认带 "/"

	os.MkdirAll(dir, os.ModeDir)
	return os.Create(dir + strFileName)
}
//删除目录或文件
func (this *FileService) DeleteFiles(fullpath string) error {
	return os.RemoveAll(fullpath)
}

func (this *FileService) GetAppUploadFileDir() string {
	return this.appRoot + "/uploads/"
}

//full path
func (this *FileService) GetUserUploadFileDir(uid string) string {
	if uid == "" {
		uid = "/default/"
	}
	return this.GetAppUploadFileDir() + uid + "/"
}

//临时目录，会生成
func (this *FileService) GetUserUploadFileDirTemp(uid string) string {
	return this.GetUserUploadFileDir(uid) //+ strconv.FormatInt(time.Now().Unix(), 10) + "/"
}

func (this *FileService) GetAllUserUploadFilePath(uid string) []string {
	return nil
}

var instanceFileService *FileService = &FileService{}
func GetFileService() *FileService {
	return instanceFileService
}

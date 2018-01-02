package main

import (
	"fmt"
	"io/ioutil"
	"os"
	//"db"
	//"github.com/golang/protobuf/proto"
	"net/http"
	//"protocol"
	//"strings"
	//"flag"
	"service"
	"handler"
	"utils"
	"time"
	"encoding/json"
	//"kLog"
	"strconv"
)

var httpInfo string = "0.0.0.0:9015"

var ServerGroups []string



func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func StartAgent(workDir string) error {
	fileService := service.GetFileService()
	fileService.SetApplicationRootDir(workDir)

	commonHandler := http.FileServer(http.Dir(workDir))
	//fmt.Println("handler ")
	http.Handle("/", commonHandler)

	//http.HandleFunc("/index", leftbarHandler)
	//http.HandleFunc("/ajax", ajaxHandler)
	http.HandleFunc("/ExecCmd", handler.ExecuteShell)
	http.HandleFunc("/UploadFileHex", handler.FileUploadHex)
	http.HandleFunc("/UploadFile", handler.FileUploadStream)
	http.HandleFunc("/ProcessDealer", handler.DealWithGetProcessRequest)
	err := http.ListenAndServe(utils.G_StAgentConf.StrHttpPort, nil)
	return err

}
func  StartHeartBeatTimer() error {
	fmt.Println("StartMysqlWriteTimer begin")
	duration := time.Second*10
	fmt.Println("duration=", duration)
	timer := time.NewTimer(duration)
	go func() {
		<-timer.C
		fmt.Println("StartTimer.time done", duration)
		go SendHeartBeat2Center()
		StartHeartBeatTimer()
	}()
	return nil
}

func  SendHeartBeat2Center()  {
	strUrl := utils.G_StAgentConf.StrCenterUrl + ""
	var HeatBeatMap map[string]string = make(map[string]string)
	HeatBeatMap["action"] = "agent_heartbeat"

	var dataMap map[string]string = make(map[string]string)
	dataMap["Innerip"] = utils.G_StAgentConf.StrInnerIp
	dataMap["Outerip"] = utils.G_StAgentConf.StrOuterIp
	dataMap["LastUpdateTime"] = strconv.FormatInt(time.Now().Unix(),10)
	dataBuff,_:= json.Marshal(dataMap)
	HeatBeatMap["data"] = string(dataBuff)
	utils.HttpPostForm(strUrl,&HeatBeatMap)
}

func loadConf() {
	fi, err := os.Open("AgentConf.json")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	fmt.Println("loadConf:", string(fd))
	err = json.Unmarshal(fd, &utils.G_StAgentConf)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(0)
	}
	return
}

//-workDir "D:/ServerAgent/"
func main() {
	// 从命令行获取参数
	loadConf()
	workDir:= &utils.G_StAgentConf.StrWorkDir
	if *workDir == "" {
		utils.Errorf("start process falied! please input workDir")
		os.Exit(1)
	}
	//ServerManager.GetFacadeZmqDealer().Init(9701)

	utils.Infof("Server starting...")
	go StartHeartBeatTimer()
	StartAgent(*workDir)
}

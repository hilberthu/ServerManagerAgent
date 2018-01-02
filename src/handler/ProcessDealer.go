package handler

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"utils"
	"encoding/json"
	"time"
)

func FunctionMapCall(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

//透传http，仅代理请求
func DealWithGetProcessRequest(w http.ResponseWriter, r *http.Request) {
	funcs := make(map[string]interface{})
	funcs["process_getserverconfig"] = DealWithGetServerConfigFile
	funcs["process_heartbeat"] = DealWithProcessHeartbeat

	strAction := r.FormValue("action")
	strData := r.FormValue("data")
	fmt.Println("ajaxHandler---------------", strAction, strData)
	FunctionMapCall(funcs, strAction, w, r, strData)

}

func DealWithGetServerConfigFile(w http.ResponseWriter, r *http.Request, strData string) {
	retMap := utils.GetRetMap()
	mapParam := make(map[string]string)
	mapParam["action"] = "process_getserverconfig"
	mapParam["data"] = strData
	bRet, strResult := utils.HttpPostForm(utils.G_StAgentConf.StrCenterUrl, &mapParam)
	fmt.Println("HttpPostForm result",utils.G_StAgentConf.StrCenterUrl,bRet,strResult)
	if bRet {
		//retMap["data"] = strResult
		json.Unmarshal([]byte(strResult),&mapParam)
		retMap["data"] = mapParam["data"]
	} else {
		retMap["ret"] = "-1"
	}
	fmt.Fprint(w, RetMap2String(retMap))
}

func DealWithProcessHeartbeat(w http.ResponseWriter, r *http.Request, strData string) {
	retMap := utils.GetRetMap()
	mapData := make(map[string]string)
	mapParam := make(map[string]string)
	mapParam["action"] = "process_heartbeat"
	err:=json.Unmarshal([]byte(strData),&mapData)
	if err!= nil{
		retMap["ret"] = "-1"
		fmt.Fprint(w, RetMap2String(retMap))
		return
	}
	mapData["Innerip"] = utils.G_StAgentConf.StrInnerIp
	mapData["Outerip"] = utils.G_StAgentConf.StrOuterIp
	mapData["Lastupdatetime"] = fmt.Sprintf("%d",time.Now().Unix())
	strSendData,_ := json.Marshal(mapData)
	mapParam["data"] = string(strSendData)
	utils.HttpPostForm(utils.G_StAgentConf.StrCenterUrl, &mapParam)
	fmt.Fprint(w, RetMap2String(retMap))
}
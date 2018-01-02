package handler

import (
	"fmt"
	"encoding/json"
	//"github.com/golang/protobuf/proto"
	"net/http"
	//"strings"
	"service"
	"strings"
	//"os"
	"strconv"
	"time"
	"utils"
)

//req.ParseForm 只能调用一次，否则出错
type BaseAction struct {

}

func GetRetMap(uid string) map[string]string {
	var retDataMap map[string]string = make(map[string]string)
	retDataMap["ret"] = "0"
	retDataMap["data"] = ""
	retDataMap["resid"] = strconv.FormatInt(time.Now().Unix(), 10)
	return retDataMap
}

func GetUID(r *http.Request) string{
	uid := r.FormValue("uid")
	if uid == "" {
		return "888"
	}
	return uid
}

func RetMap2String(retMap map[string]string) string {
	v, _ := json.Marshal(retMap)
	return string(v)
}

//http response数据
func DoResponse(w http.ResponseWriter, respContent string){
	w.Write([]byte(respContent))
}

func getParameter(r *http.Request, name string) string{
	return r.Form[name][0]
}
func getParameterList(r *http.Request, name string) []string{
	return r.Form[name] //strings.Join(v, ""))
}

func ExecuteShell(w http.ResponseWriter, r *http.Request) {
	utils.Debugf("ExecuteShell---", *r)
	r.ParseForm()//解析参数，默认是不会解析的

	uid := GetUID(r)
	cmdStr := getParameter(r, "cmd")
	utils.Debugf("commnd:", cmdStr)

	args := getParameterList(r, "args")
	//args := make([]string, 0, 1)
	_, outputLines := service.ExecCommand(cmdStr, args)

	//response
	retMap := GetRetMap(uid)
	retMap["data"] = strings.Join(outputLines, "<br/>")

	fmt.Fprint(w, RetMap2String(retMap))
}

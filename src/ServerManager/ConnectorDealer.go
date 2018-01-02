package ServerManager

import (
	//"database/sql"
	//"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	//zmq4 "github.com/pebbe/zmq4"
	"protocol"
	//"strconv"
	"time"
	//"io/ioutil"
	//"kCommon"
	//"kLog"
	//"kProtoBuf"
	//"kRpc"
	//"kServerCommon"
	//"net/http"
	//"strconv"
	//"strings"
	"sync"
	//"time"
	//"github.com/pebbe/zmq4"
)

type NodeHeartBeat struct {
	strZmqId        string
	lastRefreshTime int64
}

type ServerHeartBeat struct {
	appInfo         protocol.AppServerInfo
	lastRefreshTime int64
}

type FacadeZmqDealer struct {
	mapInnerIp2ZmqId     map[string]NodeHeartBeat
	mapInnerIp2ZmqIdLock sync.RWMutex

	mapAppInfo     map[string]map[string]ServerHeartBeat
	mapAppInfoLock sync.RWMutex

	//ZmqSoket *zmq4.Socket

	//zmq回包channel
	ZmqRspChan chan *protocol.NodeServerPkg
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (this *FacadeZmqDealer) Init(nPort int) {
	this.mapInnerIp2ZmqId = make(map[string]NodeHeartBeat)
	this.mapAppInfo = make(map[string]map[string]ServerHeartBeat)
	this.ZmqRspChan = make(chan *protocol.NodeServerPkg, 1)
	go this.ZeroMqLoop(nPort)
}

func (this *FacadeZmqDealer) AddInnerIp2Map(strInnerIp string, strZmqId string) {
	this.mapInnerIp2ZmqIdLock.Lock()
	var tempNodeHeartBeat NodeHeartBeat
	tempNodeHeartBeat.lastRefreshTime = time.Now().Unix()
	tempNodeHeartBeat.strZmqId = strZmqId
	this.mapInnerIp2ZmqId[strInnerIp] = tempNodeHeartBeat
	this.mapInnerIp2ZmqIdLock.Unlock()
	fmt.Println("AddInnerIp2Map", strInnerIp, strZmqId)
}

func (this *FacadeZmqDealer) AddAppInfo2Map(strServerName string, strInsId string, appInfo *protocol.AppServerInfo) {
	this.mapAppInfoLock.Lock()
	if _, OK := this.mapAppInfo[strServerName][strInsId]; !OK {
		this.mapAppInfo[strServerName] = make(map[string]ServerHeartBeat)
	}
	var stServerHeartBeat ServerHeartBeat
	stServerHeartBeat.lastRefreshTime = time.Now().Unix()
	stServerHeartBeat.appInfo = *appInfo
	this.mapAppInfo[strServerName][strInsId] = stServerHeartBeat
	this.mapAppInfoLock.Unlock()
	fmt.Println("AddAppInfo2Map", strServerName, strInsId, *appInfo)
}

func (this *FacadeZmqDealer) GetAppLastRefreshTime(strServerName string, strInsId string) int64 {
	this.mapAppInfoLock.Lock()
	defer this.mapAppInfoLock.Unlock()
	if oneServerHeartBeat, OK := this.mapAppInfo[strServerName][strInsId]; !OK {
		return -1
	} else {
		return oneServerHeartBeat.lastRefreshTime
	}

}

func (this *FacadeZmqDealer) InnerIp2Zmqid(strInnerIp string) string {
	this.mapInnerIp2ZmqIdLock.Lock()
	defer this.mapInnerIp2ZmqIdLock.Unlock()
	stNodeHeartBeat, OK := this.mapInnerIp2ZmqId[strInnerIp]
	if OK {
		fmt.Println("InnerIp2Zmqid ret", stNodeHeartBeat.strZmqId)
		return stNodeHeartBeat.strZmqId
	}
	return ""

}

func (this *FacadeZmqDealer) GetNodeList() []string {
	this.mapInnerIp2ZmqIdLock.Lock()
	defer this.mapInnerIp2ZmqIdLock.Unlock()
	var NodeList []string
	for strNodeId, _ := range this.mapInnerIp2ZmqId {
		NodeList = append(NodeList, strNodeId)
	}
	return NodeList

}

/*
type NodeMkNewAppReqPkg struct {
    Path             *string        `protobuf:"bytes,1,req,name=path" json:"path,omitempty"`
    CenterUrl        *string        `protobuf:"bytes,3,req,name=centerUrl" json:"centerUrl,omitempty"`
    AppInfo          *AppServerInfo `protobuf:"bytes,4,req,name=appInfo" json:"appInfo,omitempty"`
    XXX_unrecognized []byte         `json:"-"`
}
*/

func (this *FacadeZmqDealer) SendMkNewAppCmd(strNode string, strPath string, strServerName string, strInsId string, nPort int32, strOuterIp string, strOther string) int32 {
	var stpkg protocol.NodeServerPkg
	stpkg.Cmd = proto.Int32(int32(protocol.NodeCmd_NODE_CMD_MKNEW_APP))
	stpkg.Seq = proto.Uint64(uint64(time.Now().Unix())) //生成一个随机数uuid
	var stMkNewPkg protocol.NodeMkNewAppReqPkg
	stMkNewPkg.Path = proto.String(strPath)

	var stAppInfo protocol.AppServerInfo
	stAppInfo.Port = proto.Int32(nPort)
	stAppInfo.StrServerName = proto.String(strServerName)
	stAppInfo.StrInstId = proto.String(strInsId)
	stMkNewPkg.AppInfo = &stAppInfo
	stMkNewPkg.CenterUrl = proto.String("tcp://10.20.104.175:9701")
	mkNewPkgBuff, err := proto.Marshal(&stMkNewPkg)
	if err != nil {
		fmt.Println("marshaling error: ", err)
		return -1
	}
	stpkg.Data = mkNewPkgBuff

	pkgBuff, err := proto.Marshal(&stpkg)
	if err != nil {
		fmt.Println("marshaling error: ", err)
		return -1
	}

	fmt.Println("nodeServerpkg==", string(pkgBuff))
	strZmqId := this.InnerIp2Zmqid(strNode)
	if strZmqId == "" {
		fmt.Println("can't find == strNode", strNode, this.mapInnerIp2ZmqId)
		return -1
	} else {
		//this.ZmqSoket.SendBytes([]byte(strZmqId), zmq4.SNDMORE)
		//this.ZmqSoket.SendBytes(pkgBuff, 0)
		timeout := make(chan bool, 1)
		go func() {
			time.Sleep(10000e6) //等待2毫秒
			timeout <- true
		}()
		select {
		case pstZmqPkg := <-this.ZmqRspChan:
			if pstZmqPkg.GetSeq() != stpkg.GetSeq() {
				fmt.Println("pstZmqPkg.Seq error", pstZmqPkg.Seq, stpkg.Seq)
				return -1
			}
			var stNodeMkNewAppRspPkg protocol.NodeMkNewAppRspPkg
			err = proto.Unmarshal([]byte(pstZmqPkg.GetData()), &stNodeMkNewAppRspPkg)
			if err != nil {
				fmt.Println("stNodeMkNewAppRspPkg error", err.Error())
				return -1
			}
			//(this *FacadeMysql) MkNewServer(strInnerIp string, strServerName string, strInsId string, strPath string, port int, strOuterIp string, strOther string)
			nRet := stNodeMkNewAppRspPkg.GetRet()
			if nRet == 0 {
				FacadeMysql := GetFacadeMysql()
				FacadeMysql.MkNewServer(strNode, strServerName, strInsId, strPath, nPort, strOuterIp, strOther)
			}
		case <-timeout:
		}

	}

	return 0
}

func (this *FacadeZmqDealer) SendStopCmd(strNode string, strPath string, strServerName string, strInsId string) int32 {
	//var stpkg protocol.NodeServerPkg
	//stpkg.Cmd = proto.Int32(int32(protocol.NodeCmd_NODE_CMD_STOP))
	//stpkg.Seq = proto.Uint64(uint64(time.Now().Unix())) //生成一个随机数uuid
	//var stStopPkg protocol.NodeStopReqPkg
	//stStopPkg.Path = proto.String(strPath)
	//
	//var stAppInfo protocol.AppServerInfo
	//stAppInfo.Port = proto.Int32(0)
	//stAppInfo.StrServerName = proto.String(strServerName)
	//stAppInfo.StrInstId = proto.String(strInsId)
	//stAppInfo.Path = proto.String(strPath)
	//stStopPkg.AppInfo = &stAppInfo
	//mkNewPkgBuff, err := proto.Marshal(&stStopPkg)
	//if err != nil {
	//	fmt.Println("marshaling error: ", err)
	//	return -1
	//}
	//stpkg.Data = mkNewPkgBuff
	//
	//pkgBuff, err := proto.Marshal(&stpkg)
	//if err != nil {
	//	fmt.Println("marshaling error: ", err)
	//	return -1
	//}
	//
	//fmt.Println("nodeServerpkg==", string(pkgBuff))
	//strZmqId := this.InnerIp2Zmqid(strNode)
	//if strZmqId == "" {
	//	fmt.Println("can't find == strNode", strNode, this.mapInnerIp2ZmqId)
	//	return -1
	//} else {
	//	this.ZmqSoket.SendBytes([]byte(strZmqId), zmq4.SNDMORE)
	//	this.ZmqSoket.SendBytes(pkgBuff, 0)
	//	timeout := make(chan bool, 1)
	//	go func() {
	//		time.Sleep(10000e6) //等待2毫秒
	//		timeout <- true
	//	}()
	//	select {
	//	case pstZmqPkg := <-this.ZmqRspChan:
	//		if pstZmqPkg.GetSeq() != stpkg.GetSeq() {
	//			//this.ZmqRspChan <- pstZmqPkg
	//			fmt.Println("pstZmqPkg.Seq error", pstZmqPkg.GetSeq(), stpkg.GetSeq())
	//			return -1
	//		}
	//		var stRspPkg protocol.NodeStopRspPkgs
	//		err = proto.Unmarshal([]byte(pstZmqPkg.GetData()), &stRspPkg)
	//		if err != nil {
	//			fmt.Println("stRspPkg error", err.Error())
	//			return -1
	//		}
	//
	//	case <-timeout:
	//	}
	//
	//}

	return 0
}

func (this *FacadeZmqDealer) SendStartCmd(strNode string, strPath string, strServerName string, strInsId string) int32 {
	//var stpkg protocol.NodeServerPkg
	//stpkg.Cmd = proto.Int32(int32(protocol.NodeCmd_NODE_CMD_START))
	//stpkg.Seq = proto.Uint64(uint64(time.Now().Unix())) //生成一个随机数uuid
	//var stStartPkg protocol.NodeStartReqPkg
	//stStartPkg.Path = proto.String(strPath)
	//
	//var stAppInfo protocol.AppServerInfo
	//stAppInfo.Port = proto.Int32(0)
	//stAppInfo.StrServerName = proto.String(strServerName)
	//stAppInfo.StrInstId = proto.String(strInsId)
	//stAppInfo.Path = proto.String(strPath)
	//stStartPkg.AppInfo = &stAppInfo
	//mkNewPkgBuff, err := proto.Marshal(&stStartPkg)
	//if err != nil {
	//	fmt.Println("marshaling error: ", err)
	//	return -1
	//}
	//stpkg.Data = mkNewPkgBuff
	//
	//pkgBuff, err := proto.Marshal(&stpkg)
	//if err != nil {
	//	fmt.Println("marshaling error: ", err)
	//	return -1
	//}
	//
	//fmt.Println("nodeServerpkg==", string(pkgBuff))
	//strZmqId := this.InnerIp2Zmqid(strNode)
	//if strZmqId == "" {
	//	fmt.Println("can't find == strNode", strNode, this.mapInnerIp2ZmqId)
	//	return -1
	//} else {
	//	this.ZmqSoket.SendBytes([]byte(strZmqId), zmq4.SNDMORE)
	//	this.ZmqSoket.SendBytes(pkgBuff, 0)
	//	timeout := make(chan bool, 1)
	//	go func() {
	//		time.Sleep(100000e6) //等待10秒
	//		timeout <- true
	//	}()
	//	select {
	//	case pstZmqPkg := <-this.ZmqRspChan:
	//		if pstZmqPkg.GetSeq() != stpkg.GetSeq() {
	//			fmt.Println("pstZmqPkg.Seq error", pstZmqPkg.GetSeq(), stpkg.GetSeq())
	//			return -1
	//		}
	//		var stRspPkg protocol.NodeStartRspPkgs
	//		err = proto.Unmarshal([]byte(pstZmqPkg.GetData()), &stRspPkg)
	//		if err != nil {
	//			fmt.Println("stRspPkg error", err.Error())
	//			return -1
	//		}
	//
	//	case <-timeout:
	//	}
	//
	//}

	return 0
}

func (this *FacadeZmqDealer) ZeroMqLoop(nPort int) {
	//context, _ := zmq4.NewContext()
	//this.ZmqSoket, _ = context.NewSocket(zmq4.ROUTER)
	//this.ZmqSoket.SetSndhwm(10000)
	//this.ZmqSoket.SetRcvhwm(10000)
	//var err error
	//
	//if nPort > 0 {
	//	err = this.ZmqSoket.Bind("tcp://0.0.0.0:" + strconv.Itoa(nPort))
	//	fmt.Println("ZeroMqLoop Begin Listen >>>>>>>>>"+"tcp://0.0.0.0:"+strconv.Itoa(nPort), err)
	//	if nil != err {
	//		return
	//	}
	//	for {
	//		bHaveEvent := false
	//		msgId, err := this.ZmqSoket.Recv(0)
	//		var strId string
	//		if err == nil {
	//			strId = string(msgId)
	//			fmt.Println("Got id", string(strId))
	//		}
	//		msgData, err := this.ZmqSoket.Recv(0)
	//		if err == nil {
	//			fmt.Println("Got data", string(msgData))
	//		}
	//		//var ZeroMqData kProtoBuf.ZeroMqData
	//		var nodeServerpkg protocol.NodeServerPkg
	//		err = proto.Unmarshal([]byte(msgData), &nodeServerpkg)
	//		if err == nil {
	//			this.DealWithZmqMsg(this.ZmqSoket, &nodeServerpkg, &strId)
	//			bHaveEvent = true
	//		} else {
	//			fmt.Println("nodeServerpkg error", err.Error())
	//		}
	//		if !bHaveEvent {
	//			time.Sleep(time.Second / 100)
	//		}
	//	}
	//
	//}
	//fmt.Println("ZeroMqLoop Process Over >>>>>>>> ")

}

//func (this *FacadeZmqDealer) DealWithZmqMsg(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	switch pstPkg.GetCmd() {
//	//node节点的心跳
//	case int32(protocol.NodeCmd_NODE_CMD_NODE2CENTER_HEARTBEAT):
//		{
//			return this.DealWithNodeHeartBeat(s, pstPkg, strId)
//		}
//	case int32(protocol.NodeCmd_NODE_CMD_MKNEW_APP_RSP):
//		{
//			return this.DealWithMkNewAppRsp(s, pstPkg, strId)
//		}
//	case int32(protocol.NodeCmd_NODE_CMD_SERVER2CENTER_HEARTBEAT):
//		{
//			this.DealWithServerHeartBeat(s, pstPkg, strId)
//		}
//	case int32(protocol.NodeCmd_NODE_CMD_STOP_RSP):
//		{
//			this.DealWithStopRsp(s, pstPkg, strId)
//		}
//	case int32(protocol.NodeCmd_NODE_CMD_START_RSP):
//		{
//			this.DealWithStartRsp(s, pstPkg, strId)
//		}
//	}
//
//	return 0
//
//}
//
//func (this *FacadeZmqDealer) DealWithStartRsp(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	fmt.Println("DealWithStartRsp begin >>>>>>>> ")
//	this.ZmqRspChan <- pstPkg
//	fmt.Println("DealWithStartRsp end >>>>>>>> ")
//	return 0
//}
//
//func (this *FacadeZmqDealer) DealWithStopRsp(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	fmt.Println("DealWithStopRsp begin >>>>>>>> ")
//	this.ZmqRspChan <- pstPkg
//	fmt.Println("DealWithStopRsp end >>>>>>>> ")
//	return 0
//}
//
//func (this *FacadeZmqDealer) DealWithNodeHeartBeat(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	var heartBeatPkg protocol.Node2CenterHeartBeatReqPkg
//	err := proto.Unmarshal(pstPkg.GetData(), &heartBeatPkg)
//
//	if err == nil {
//		this.AddInnerIp2Map(heartBeatPkg.GetStrInnerip(), *strId)
//
//	} else {
//		fmt.Println("err==", err.Error())
//	}
//
//	return 0
//}
//
//func (this *FacadeZmqDealer) DealWithServerHeartBeat(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	var heartBeatPkg protocol.Node2CenterHeartBeatReqPkg
//	err := proto.Unmarshal(pstPkg.GetData(), &heartBeatPkg)
//
//	if err == nil {
//		appInfo := heartBeatPkg.GetAppInfo()
//		this.AddAppInfo2Map(appInfo.GetStrServerName(), appInfo.GetStrInstId(), appInfo)
//		//UpdateServerInfo(strInnerIp string, strServerName string, strInsId string, strPath string, port int32, status int)
//		FacadeMysql := GetFacadeMysql()
//
//		FacadeMysql.ServerDiscovery(heartBeatPkg.GetStrInnerip(), appInfo.GetStrServerName(), appInfo.GetStrInstId(), appInfo.GetPath(), appInfo.GetPort(), int(protocol.ServerStatus_SERVER_STATUS_RUNNING))
//
//	} else {
//		fmt.Println("err==", err.Error())
//	}
//
//	return 0
//}
//
//func (this *FacadeZmqDealer) DealWithMkNewAppRsp(s *zmq4.Socket, pstPkg *protocol.NodeServerPkg, strId *string) int {
//	fmt.Println("DealWithMkNewAppRsp begin >>>>>>>> ")
//	this.ZmqRspChan <- pstPkg
//	fmt.Println("DealWithMkNewAppRsp end >>>>>>>> ")
//	return 0
//}

var G_FacadeZmqDealer *FacadeZmqDealer

func GetFacadeZmqDealer() *FacadeZmqDealer {
	if G_FacadeZmqDealer == nil {
		G_FacadeZmqDealer = new(FacadeZmqDealer)
	}
	return G_FacadeZmqDealer
}

package ServerManager

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"protocol"
	"time"
)

type FacadeMysql struct {
	strDbConn string
	dbconn    *sql.DB
}

type ServerItem struct {
	Innerip        string
	Servername     string
	Insid          string
	Path           string
	Port           int32
	Status         int32
	Lastupdatetime string
	Outerip        string
	Other          string
	StatusString   string
}

func (this *FacadeMysql) Init(strconn string) int {

	dbconn, err := sql.Open("mysql", strconn)
	checkErr(err)
	this.strDbConn = strconn
	this.dbconn = dbconn
	return 0
}

func (this *FacadeMysql) UpdateLastTime(strServerName string, strInsId string, nTime uint64) int {
	stmt, err := this.dbconn.Prepare("update serverlist set lastupdatetime=? where servername=? and insid=?")
	checkErr(err)

	res, err := stmt.Exec(nTime, strServerName, strInsId)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)
	fmt.Println("affect==", affect)
	return 0
}

/*
CREATE TABLE `serverlist` (
  `innerip` varchar(32) NOT NULL,
  `servername` varchar(64) NOT NULL,
  `insid` varchar(8) NOT NULL,
  `path` varchar(256) NOT NULL,
  `port` int(11) NOT NULL,
  `status` int(11) NOT NULL,
  `lastupdatetime` int(64) NOT NULL,
  `outerip` varchar(32) NOT NULL,
  `other` varchar(156) DEFAULT NULL,
  PRIMARY KEY (`insid`,`servername`,`innerip`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8
*/

func (this *FacadeMysql) MkNewServer(strInnerIp string, strServerName string, strInsId string, strPath string, port int32, strOuterIp string, strOther string) int {
	stmt, err := this.dbconn.Prepare("insert into serverlist values(?,?,?,?,?,?,?,?,?)")
	checkErr(err)

	res, err := stmt.Exec(strInnerIp, strServerName, strInsId, strPath, port, 1, time.Now().Unix(), strOuterIp, strOther)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)
	fmt.Println("affect==", affect)
	return 0
}

//server自发现
func (this *FacadeMysql) ServerDiscovery(strInnerIp string, strServerName string, strInsId string, strPath string, port int32, status int) int {

	bIsExists := this.CheckIsServerExists(strServerName, strInsId)
	if bIsExists {
		fmt.Println("Server is exist in db", strServerName, strInsId)
		return 0
	}

	stmt, err := this.dbconn.Prepare("insert into serverlist values(?,?,?,?,?,?,?,?,?)")

	res, err := stmt.Exec(strInnerIp, strServerName, strInsId, strPath, port, status, time.Now().Unix(), "", "")
	if err != nil {
		fmt.Println("Server is exist in db", strServerName, strInsId, err.Error())
		return 0
	}

	affect, err := res.RowsAffected()
	fmt.Println("affect==", affect)
	return 0
}

func (this *FacadeMysql) GetServerListByServerName(strServerName string) []ServerItem {
	stmt, err := this.dbconn.Prepare("select * from  serverlist where servername=?")
	checkErr(err)
	var retList []ServerItem
	rows, err := stmt.Query(strServerName)
	now := time.Now().Unix()
	if err == nil {
		for rows.Next() {
			var stOneServerItem ServerItem
			err = rows.Scan(&stOneServerItem.Innerip, &stOneServerItem.Servername, &stOneServerItem.Insid, &stOneServerItem.Path, &stOneServerItem.Port, &stOneServerItem.Status, &stOneServerItem.Lastupdatetime, &stOneServerItem.Outerip, &stOneServerItem.Other)
			if err == nil {
				stOneServerItem.StatusString = "running"
				if (now - GetFacadeZmqDealer().GetAppLastRefreshTime(stOneServerItem.Servername, stOneServerItem.Insid)) > 60 {
					stOneServerItem.Status = int32(protocol.ServerStatus_SERVER_STATUS_OFFLINE)
					stOneServerItem.StatusString = "down"
				}
				retList = append(retList, stOneServerItem)
			}
		}
	}
	return retList
}

func (this *FacadeMysql) GetAllServerName() []string {
	stmt, err := this.dbconn.Prepare("select DISTINCT servername from  serverlist")
	checkErr(err)
	var retList []string
	rows, err := stmt.Query()
	if err == nil {
		for rows.Next() {
			var ServerName string
			err = rows.Scan(&ServerName)
			if err == nil {
				retList = append(retList, ServerName)
			}
		}
	}
	return retList
}

func (this *FacadeMysql) CheckIsServerExists(strServerName string, strInsId string) bool {
	stmt, err := this.dbconn.Prepare("select * from  serverlist where servername=? and insid=?")
	checkErr(err)
	var retList []ServerItem
	rows, err := stmt.Query(strServerName, strInsId)
	if err == nil {
		for rows.Next() {
			var stOneServerItem ServerItem
			err = rows.Scan(&stOneServerItem.Innerip, &stOneServerItem.Servername, &stOneServerItem.Insid, &stOneServerItem.Path, &stOneServerItem.Port, &stOneServerItem.Status, &stOneServerItem.Lastupdatetime, &stOneServerItem.Outerip, &stOneServerItem.Other)
			if err == nil {
				retList = append(retList, stOneServerItem)
			} else {
				fmt.Println("Scan error", err.Error())
			}
		}
	} else {
		fmt.Println("stmt.Query error", err.Error())
	}
	if len(retList) == 0 {
		return false
	}
	return true
}

var G_FacadeMysql *FacadeMysql = nil

func GetFacadeMysql() *FacadeMysql {
	if G_FacadeMysql == nil {
		G_FacadeMysql = new(FacadeMysql)
		G_FacadeMysql.Init("kdb:kingsoft123@tcp(10.20.104.175:3306)/kwebtool")
	}
	return G_FacadeMysql
}

/*func main() {
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/Test?charset=utf8")
	checkErr(err)

	//插入数据
	stmt, err := db.Prepare("INSERT userinfo SET username=?,departname=?,created=?")
	checkErr(err)

	res, err := stmt.Exec("码农", "研发部门", "2016-03-06")
	checkErr(err)

	id, err := res.LastInsertId()
	checkErr(err)

	fmt.Println(id)
	//更新数据
	stmt, err = db.Prepare("update userinfo set username=? where uid=?")
	checkErr(err)

	res, err = stmt.Exec("码农二代", id)
	checkErr(err)

	affect, err := res.RowsAffected()
	checkErr(err)

	fmt.Println(affect)

	//查询数据
	rows, err := db.Query("SELECT * FROM userinfo")
	checkErr(err)

	for rows.Next() {
		var uid int
		var username string
		var department string
		var created string
		err = rows.Scan(&uid, &username, &department, &created)
		checkErr(err)
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(department)
		fmt.Println(created)
	}

	//删除数据
	stmt, err = db.Prepare("delete from userinfo where uid=?")
	checkErr(err)

	res, err = stmt.Exec(id)
	checkErr(err)

	affect, err = res.RowsAffected()
	checkErr(err)

	fmt.Println(affect)

	db.Close()

}
*/

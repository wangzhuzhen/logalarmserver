package main

import (
	"net/http"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/routers"
	"flag"
	"github.com/golang/glog"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/mysql"
/*
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"net/url"
	"github.com/wangzhuzhen/logalarmserver/utils"
	"github.com/wangzhuzhen/logalarmserver/mysql"
	"github.com/wangzhuzhen/logalarmserver/types"
*/
)

/*
var DB = &sql.DB{}

func init(){

	var c utils.Conf
	var err error
	_, err=c.GetConf()
	if err !=nil {
		glog.Error("Failed to connect Mysql due to can not read connect configuration")
		panic(err)
	}

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&loc=%s&parseTime=true", c.MysqlUser, c.MysqlPassword, c.MysqlHost, c.MysqlPort, url.QueryEscape("Asia/Shanghai"))
	DB, err := sql.Open("mysql", uri)
	if err != nil {
		glog.Error("Failed to onnect Mysql due to open Mysql")
		panic(err)
	}

	// 连接Mysql成功时确保数据库已经成功创建，如果未成功创建，则接下来的表操作也无法进行
	ret, err:= mysql.CreteDatabase(DB, types.DBname); if !ret{
		glog.Error("Falied to verify the existences database %s in Mysql", types.DBname)
		panic(err)
	}

	//db,_ = sql.Open("mysql", "root:root@/book")
}


*/


func main() {
	/* 初始化命令行参数 */
	flag.Parse()
	/* 退出时调用，确保日志写入文件中 */
	defer glog.Flush()

	glog.Info("Starting Logalarm Server.......")
	if !mysql.DB_Initial(){
		glog.Error("Can not initializing DB")
		return
	}


	router := routers.NewRouter()
	glog.Fatal(http.ListenAndServe(":8989", router))
}


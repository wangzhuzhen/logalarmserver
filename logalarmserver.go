package main

import (
	"net/http"
	"github.com/wangzhuzhen/logalarmserver/routers"
	"flag"
	"github.com/golang/glog"
)

func main() {
	/* 初始化命令行参数 */
	flag.Parse()
	/* 退出时调用，确保日志写入文件中 */
	defer glog.Flush()

	glog.Info("Starting Logalarm Server.......")

	router := routers.NewRouter()
	glog.Fatal(http.ListenAndServe(":8989", router))
}



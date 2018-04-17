package utils

import (
	"os"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/types"
	"strings"
	"github.com/golang/glog"
	"strconv"
	"time"
)

// 记录用户操作日志,传入用户id和具体操作内容
func RecordUserOperations(userId int, str_content string)  {
	err := os.MkdirAll(types.UserOperationLogDir, 0777)
	if err != nil {
		glog.Error(err)

	}

	operation_time:=time.Now().Format("2006-01-02 15:04:05");
	userLog := "{\"module\":\"日志报警\",\"userId\":" + strconv.Itoa(userId) +""+ ",\"operation\":\""+ str_content + "\",\"time\":\"" + operation_time +"\"}"

	fd,_:=os.OpenFile(types.UserOperationLogDir + "/" + types.UserOperationLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	fd_content:=strings.Join([]string{userLog,"\n"},"")
	buf:=[]byte(fd_content)
	fd.Write(buf)
	fd.Close()
}

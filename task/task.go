package task

import (
	"bytes"
	"os/exec"
	"github.com/golang/glog"
)


/* 调用系统指令执行 shell 命令 */
func RunTask(s string) bool {
	cmd := exec.Command("/bin/bash", "-c", s)     /* 调用Command函数 */
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		glog.Fatal(err)
		return false
	}
	glog.Infof("Task Running Output is:\n %s", out.String())     /*输出执行结果 */
	return true
}


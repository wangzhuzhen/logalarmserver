package task

import (
"bytes"
"fmt"
"os/exec"
"log"
)


//调用系统指令的方法，参数s 就是调用的shell命令
func RunTask(s string) bool {
	cmd := exec.Command("/bin/bash", "-c", s) //调用Command函数
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		return false
	}
	fmt.Printf("%s", out.String()) //输出执行结果
	return true
}


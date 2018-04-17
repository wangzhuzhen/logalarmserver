package utils

import (
	"github.com/golang/glog"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/task"
)

func TopologySubmit(cmd string, taskName string) bool {
	glog.Infof("Submit Topology task %s with CMD[%s]", taskName, cmd)
	//StringCMDTest := "echo test submit task"
	if !task.RunTask(cmd) {
		glog.Errorf("Exec Submit Topology Task CMD [%s] Failed", cmd)
		return false
	}
	return true
}
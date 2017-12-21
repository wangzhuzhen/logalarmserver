package handler

import (
	"fmt"
	"net/http"
	"github.com/wangzhuzhen/logalarmserver/types"
	"encoding/json"
	"io/ioutil"
	"io"
	"github.com/wangzhuzhen/logalarmserver/mysql"
	"github.com/wangzhuzhen/logalarmserver/task"
	"strings"
	"strconv"
	"github.com/golang/glog"
)

/* 服务状态检查 */
func StatusCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to Logalarm server!")
	glog.Info("Curl StatusCheck API to check the Status of Logalarm Server")
}

/* 查看所有报警规则 */
func ListRules(w http.ResponseWriter, r *http.Request) {

	glog.Info("Submit ListRules request to Logalarm Server")
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in ListRules()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	rules, err := mysql.Select_Rules(db, types.DBname)
	if err != nil {
		glog.Errorf("Select data for all users in database [%s] failed in ListRules()", types.DBname)
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	glog.Info("Successful processing the ListRules request to Logalarm Server")
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警规则 */
func ListUserRules(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	user := paths[len(paths) - 1]

	glog.Infof("Submit ListUserRules [User= %s] request to Logalarm Server", user)
	if (len(paths) != 3) {
		glog.Error("URI.Path is invalid in ListUerRules(). Expected is /rules/{ruleOwner}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in ListUserRules()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	rules, err := mysql.Select_UserRules(db, types.DBname, user)
	if err != nil {
		glog.Errorf("Select data for user [%s] in database [%s] failed in ListUserRules()", user, types.DBname)
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	if (rules == nil) {
		glog.Infof("No matched rules for owner [%s] or owner [%s] is non-existent in database [%s]", user, user, types.DBname)
		glog.Infof("Successful processing the ListUserRules [User= %s] request to Logalarm Server", user)
		HttpResponse(w, http.StatusNoContent)
		return
	}

	glog.Infof("Successful processing the ListUserRules [User= %s] request to Logalarm Server", user)
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 创建新的报警规则 */
func CreateRule(w http.ResponseWriter, r *http.Request) {
	glog.Info("Submit CreateRule request to Logalarm Server")
	var rule types.Rule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		glog.Error("Unmarshal the request body failed in CreateRule()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in CreateRule()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)
	mysql.CreteTable_Rule(db, types.DBname)
	mysql.Insert_Rule(db, types.DBname,rule)

	glog.Info("Successful processing the CreateRule request to Logalarm Server")
	HttpResponse(w, http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
}

/* 更新指定用户指定规则名称的报警规则 */
func UpdateRule(w http.ResponseWriter, r *http.Request) {

	var rule types.Rule
	paths := strings.Split(r.URL.Path, "/")
	rule.RuleOwner = paths[len(paths) - 2]
	rule.RuleName = paths[len(paths) - 1]
	glog.Infof("Submit UpdateRule [RuleOwner=%s, RuleName=%s] request to Logalarm Server", rule.RuleOwner, rule.RuleName)
	if (len(paths) != 4) {
		glog.Error("URI.Path is invalid. Expected is /rules/{ruleOwner}/{ruleName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		glog.Error("Unmarshal the request body failed in UpdateRule()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in UpdateRule()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.Update_Rule(db, types.DBname, rule)

	glog.Info("Successful processing the UpdateRule request to Logalarm Server")
	HttpResponse(w, http.StatusResetContent)
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
}

/* 删除指定用户指定规则名称的报警规则 */
func DeleteRule(w http.ResponseWriter, r *http.Request) {

	var drule types.RuleUser
	paths := strings.Split(r.URL.Path, "/")
	drule.RuleOwner = paths[len(paths) - 2]
	drule.RuleName = paths[len(paths) - 1]
	glog.Infof("Submit DeleteRule [RuleOwner=%s, RuleName=%s] request to Logalarm Server", drule.RuleOwner, drule.RuleName)
	if (len(paths) != 4) {
		glog.Error("URI.Path is invalid. Expected is /rules/{ruleOwner}/{ruleName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in DeleteRule()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	glog.Info("Successful processing the DeleteRule request to Logalarm Server")
	mysql.DeleteRule(db, types.DBname, drule)
	HttpResponse(w, http.StatusOK)
}

/* 查看所有报警任务 */
func ListTopologys(w http.ResponseWriter, r *http.Request) {

	glog.Info("Submit ListTopologys request to the Logalarm Server")
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in ListTopologys()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err := mysql.Select_Topologys(db, types.DBname)
	if err != nil {
		glog.Errorf("Select data for database [%s] failed in ListTopologys()", types.DBname)
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	glog.Info("Successful processing the ListTopologys request to Logalarm Server")
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(topologys); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警任务 */
func ListUserTopologys(w http.ResponseWriter, r *http.Request) {

	var user string
	paths := strings.Split(r.URL.Path, "/")
	user =  paths[len(paths) - 1]
	glog.Infof("Submit ListUserTopologys [User=%s] request to Logalarm Server", user)
	if (len(paths) != 3) {
		glog.Error("URI.Path is invalid. Expected is /topologys/{topologyOwner}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	if paths[len(paths) - 1] == "submit" {
		glog.Error("URI.Path is invalid. No User [submit] is defined")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in ListUserTopologys()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err := mysql.Select_UserTopologys(db, types.DBname, user)
	if err != nil {
		glog.Errorf("Select data for user [%s] in database [%s] failed in ListUserTopologys()", user, types.DBname)
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	if (topologys == nil) {
		glog.Infof("No matched topologys in table topologys for owner [%s] or owner [%s] is non-existent", user, user)
		HttpResponse(w, http.StatusNoContent)
		return
	}

	glog.Info("Successful processing the ListUserTopologys request to Logalarm Server")
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(topologys); err != nil {
		panic(err)
	}
}

/* 创建新的报警任务记录（已弃用，任务表的删除创建数据项目都只能通过任务提交和删除的操作入口处理） */
/*
func CreateTopology(w http.ResponseWriter, r *http.Request) {

	var topology types.Topology
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 4194304))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(body, &topology); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(topology); err != nil {
		panic(err)
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)

	mysql.CreteTable_Topology(db, types.DBname)
	mysql.Insert_Topology(db, types.DBname,topology)
	//TODO：此处提交报警任务
}
*/

/* 更新指定用户指定任务名称的报警任务,同时删除旧任务，提交新任务 */
func UpdateTopology(w http.ResponseWriter, r *http.Request) {

	var topology types.Topology
	paths := strings.Split(r.URL.Path, "/")
	topology.TopologyName = paths[len(paths) - 1]
	topology.TopologyOwner = paths[len(paths) - 2]
	glog.Infof("Submit UpdateTopology request [TopologyOwner=%s, TopologyName=%s] to Logalarm Server", topology.TopologyOwner, topology.TopologyName)
	if (len(paths) != 4) {
		glog.Error("URI.Path is invalid. Expected is /topologys/{topologyOwner}/{topologyName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := json.Unmarshal(body, &topology); err != nil {
		glog.Error("Unmarshal the request body failed in UpdateTopology()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	/* 此处杀掉正在运行的任务，重新提交任务 */
	/* 杀掉任务 */
	StringCMD := "/opt/strom/bin/storm kill " + topology.TopologyName + " -w 30"
	glog.Infof("The Topology task Kill CMD is: %s", StringCMD)
	StringCMDTest := "echo test kill task"
	if !task.RunTask(StringCMDTest) {
		glog.Errorf("Exec Update Kill Topology Task CMD [%s] Failed", StringCMD)
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

	/* 更新任务表数据 */
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in UpdateTopology()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	mysql.Update_Topology(db, types.DBname, topology)

	/* 重新提交任务 */
	updatedTaskCMD := "/opt/strom/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + topology.TopologyName + " " + topology.AppName +
		" " + strconv.Itoa(topology.KeywordIndex) + " " + topology.KeyWord + " " + strconv.Itoa(topology.TimeWindow) + " " + strconv.Itoa(topology.ThresholdNum) + " " + topology.EmailList
	glog.Infof("Update Topology task CMD is: " + updatedTaskCMD)
	updatedTaskToRun := "echo test submit update task"
	if !task.RunTask(updatedTaskToRun){
		glog.Errorf("Exec Update Submit Topology Task CMD [%s] Failed", updatedTaskCMD)
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

	glog.Info("Successful processing the UpdateTopology request to Logalarm Server")
	HttpResponse(w, http.StatusResetContent)
	if err := json.NewEncoder(w).Encode(topology); err != nil {
		panic(err)
	}
}

/* 删除指定用户指定报警任务名称的报警任务同时删除任务 (已弃用，任务表的删除创建数据项目都只能通过任务提交和删除的操作入口处理) */
/*
func DeleteTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyOwner}/{topologyName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	var dtopology types.TopologyUser

	dtopology.TopologyOwner = paths[len(paths) - 2]
	dtopology.TopologyName = paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	// 杀掉报警任务
	StringCMD := "/opt/strom/bin/storm kill " + dtopology.TopologyName + " -w 30"
	fmt.Println("StringCMD is: " + StringCMD)
	StringCMDTest := "echo  kill task due to delete topology in table topologys"
	if !task.RunTask(StringCMDTest) {
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}
	mysql.DeleteTopology(db, types.DBname, dtopology)

	HttpResponse(w, http.StatusOK)

}
*/

/* 根据任务名称提交报警任务 */
func SubmitTopology(w http.ResponseWriter, r *http.Request) {

	glog.Info("Submit SubmitTopology request to Logalarm Server")
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) != 3 {
		glog.Error("URI.Path is invalid. Expected is /topologys/submit")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	if paths[len(paths) - 1] != "submit" {
		glog.Error("URI.Path is invalid. Expected end substring of URI.Path is [submit]")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	var topology types.Topology
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	if err := json.Unmarshal(body, &topology); err != nil {
		glog.Error("Unmarshal the request body failed in SubmitTopology()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}
	if topology.TopologyOwner == "submit" {
		glog.Warning("Can not submit a Topology with TopologyOwner=submit")
		return
	}

	StringCMD := "/opt/strom/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + topology.TopologyName + " " + topology.AppName +
		" " + strconv.Itoa(topology.KeywordIndex) + " " + topology.KeyWord + " " + strconv.Itoa(topology.TimeWindow) + " " + strconv.Itoa(topology.ThresholdNum) + " " + topology.EmailList
	glog.Infof("Submit Topology task CMD is: " + StringCMD)
	StringCMDTest := "echo test submit task"
	if !task.RunTask(StringCMDTest){
		glog.Errorf("Exec Submit Topology Task CMD [%s] Failed", StringCMD)
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in SubmitTopology()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)
	mysql.CreteTable_Topology(db, types.DBname)
	mysql.Insert_Topology(db, types.DBname,topology)

	glog.Info("Successful processing the SubmitTopology request to Logalarm Server")
	HttpResponse(w, http.StatusCreated)
	if err := json.NewEncoder(w).Encode(topology); err != nil {
		panic(err)
	}
}

/* 根据任务名称提交报警任务 */
func KillTopology(w http.ResponseWriter, r *http.Request) {

	var taskOwner types.TopologyUser
	paths := strings.Split(r.URL.Path, "/")
	taskOwner.TopologyOwner = paths[len(paths) - 3]
	taskOwner.TopologyName = paths[len(paths) - 2]
	glog.Infof("Submit KillTopology request [TopologyOwner=%s, TopologyName=%s] to Logalarm Server", taskOwner.TopologyOwner, taskOwner.TopologyName)
	if len(paths) != 5  {
		glog.Error("URI.Path is invalid. Expected is /topologys/{topologyOwner}/{topologyName}/kill")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		glog.Error("Connect to Mysql failed in KillTopology()")
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err :=mysql.Select_SubmitTopology(db, types.DBname, taskOwner)
	if err != nil {
		glog.Error("Error to read topology task from topologys table")
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}
	if topologys == nil {
		glog.Infof("No matched topology task [topologyOwner=%s, topologyName=%s] in table topologys", taskOwner.TopologyOwner, taskOwner.TopologyName)
		HttpResponse(w, http.StatusNoContent)
		return
	}

	for i:=0;i<len(topologys);i++ {
		taskName := topologys[i].TopologyName

		StringCMD := "/opt/strom/bin/storm kill " + taskName + " -w 30"
		glog.Infof("The Topology task Kill CMD is: %s", StringCMD)
		StringCMDTest := "echo test kill task"
		if !task.RunTask(StringCMDTest){
			glog.Errorf("Exec Kill Topology Task CMD [%s] Failed", StringCMD)
			HttpResponse(w, http.StatusUnprocessableEntity)
			return
		}

		/* 杀掉任务后，再将topologys 表中对应的任务项删除 */
		mysql.DeleteTopology(db, types.DBname, taskOwner)
	}

	glog.Info("Successful processing the KillTopology request to the Logalarm Server")
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(taskOwner); err != nil {
		panic(err)
	}
}

func HttpResponse(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

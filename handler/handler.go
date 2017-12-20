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
)

/* 服务状态检查 */
func StatusCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to Logalarm server!")
}

/* 查看所有报警规则 */
func ListRules(w http.ResponseWriter, r *http.Request) {

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	rules, err := mysql.Select_Rules(db, types.DBname)
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	HttpResponse(w, http.StatusOK)

	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警规则 */
func ListUserRules(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 3) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleOwner}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	var user string
	user = paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	rules, err := mysql.Select_UserRules(db, types.DBname, user)
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	if (rules == nil) {
		fmt.Println("No matched rules for owner [" + user +"] " + "or owner is non-existent")
		HttpResponse(w, http.StatusNoContent)
		return
	}

	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 创建新的报警规则 */
func CreateRule(w http.ResponseWriter, r *http.Request) {
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
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)

	mysql.CreteTable_Rule(db, types.DBname)
	mysql.Insert_Rule(db, types.DBname,rule)


	HttpResponse(w, http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
}

/* 更新指定用户指定规则名称的报警规则 */
func UpdateRule(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleOwner}/{ruleName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

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
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	rule.RuleOwner = paths[len(paths) - 2]
	rule.RuleName = paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.Update_Rule(db, types.DBname, rule)

	HttpResponse(w, http.StatusResetContent)
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
}

/* 删除指定用户指定规则名称的报警规则 */
func DeleteRule(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleOwner}/{ruleName}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	var drule types.RuleUser
	drule.RuleOwner = paths[len(paths) - 2]
	drule.RuleName = paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.DeleteRule(db, types.DBname, drule)
	HttpResponse(w, http.StatusOK)
}

/* 查看所有报警任务 */
func ListTopologys(w http.ResponseWriter, r *http.Request) {

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err := mysql.Select_Topologys(db, types.DBname)
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(topologys); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警任务 */
func ListUserTopologys(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 3) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyOwner}")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	if paths[len(paths) - 1] == "submit" {
		fmt.Println("URI.Path is invalid. The topologyOwner [submit] is forbidden")
		HttpResponse(w, http.StatusNotFound)
		return
	}
	var user string
	user =  paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err := mysql.Select_UserTopologys(db, types.DBname, user)
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	if (topologys == nil) {
		fmt.Println("No matched topologys in table topologys for owner [" + user +"] " + "or owner is non-existent")
		HttpResponse(w, http.StatusNoContent)
		return
	}

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

/* 更新指定用户指定任务名称的报警任务 */
func UpdateTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyOwner}/{topologyName}")
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
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	topology.TopologyName = paths[len(paths) - 1]
	topology.TopologyOwner = paths[len(paths) - 2]

	/* 此处杀掉正在运行的任务，重新提交任务 */
	/* 杀掉任务 */
	StringCMD := "/opt/strom/bin/storm kill " + topology.TopologyName + " -w 30"
	fmt.Println("StringCMD is: " + StringCMD)
	StringCMDTest := "echo test kill task"
	if !task.RunTask(StringCMDTest) {
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

	/* 更新任务表数据 */
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}
	mysql.Update_Topology(db, types.DBname, topology)

	/* 重新提交任务 */
	updatedTask := "/opt/strom/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + topology.TopologyName + " " + topology.AppName +
		" " + strconv.Itoa(topology.KeywordIndex) + " " + topology.KeyWord + " " + strconv.Itoa(topology.TimeWindow) + " " + strconv.Itoa(topology.ThresholdNum) + " " + topology.EmailList
	fmt.Println("Updated task is: " + updatedTask)
	updatedTaskToRun := "echo test submit update task"
	if !task.RunTask(updatedTaskToRun){
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

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

	paths := strings.Split(r.URL.Path, "/")
	if len(paths) != 3 {
		fmt.Println("URI.Path is invalid. Expected is /topologys/submit")
		HttpResponse(w, http.StatusNotFound)
		return
	}

	if paths[len(paths) - 1] != "submit" {
		fmt.Println("URI.Path is invalid. Expected end substring of URI.Path is [submit]")
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
		HttpResponse(w, http.StatusUnprocessableEntity)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	StringCMD := "/opt/strom/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + topology.TopologyName + " " + topology.AppName +
		" " + strconv.Itoa(topology.KeywordIndex) + " " + topology.KeyWord + " " + strconv.Itoa(topology.TimeWindow) + " " + strconv.Itoa(topology.ThresholdNum) + " " + topology.EmailList
	fmt.Println("StringCMD is: " + StringCMD)
	StringCMDTest := "echo test submit task"
	if !task.RunTask(StringCMDTest){
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)
	mysql.CreteTable_Topology(db, types.DBname)
	mysql.Insert_Topology(db, types.DBname,topology)

	HttpResponse(w, http.StatusCreated)
	if err := json.NewEncoder(w).Encode(topology); err != nil {
		panic(err)
	}
}

/* 根据任务名称提交报警任务 */
func KillTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if len(paths) != 5  {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyOwner}/{topologyName}/kill")
		HttpResponse(w, http.StatusNotFound)
		return
	}
	var taskOwner types.TopologyUser
	taskOwner.TopologyOwner = paths[len(paths) - 3]
	taskOwner.TopologyName = paths[len(paths) - 2]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		HttpResponse(w, http.StatusUnprocessableEntity)
		panic(err)
	}

	topologys, err :=mysql.Select_SubmitTopology(db, types.DBname, taskOwner)
	if err != nil {
		fmt.Println("Error to read topology task from topologys table")
		HttpResponse(w, http.StatusUnprocessableEntity)
		return
	}
	if topologys == nil {
		fmt.Println("No matched topology task [topologyOwner=" + taskOwner.TopologyOwner + " topologyName=" +taskOwner.TopologyName + "] in table topologys")
		HttpResponse(w, http.StatusNoContent)
		return
	}

	for i:=0;i<(len(topologys));i++ {
		taskName := topologys[i].TopologyName

		StringCMD := "/opt/strom/bin/storm kill " + taskName + " -w 30"
		fmt.Println("StringCMD is: " + StringCMD)
		StringCMDTest := "echo test kill task"
		if !task.RunTask(StringCMDTest){
			HttpResponse(w, http.StatusUnprocessableEntity)
			return
		}

		/* 杀掉任务后，再将topologys 表中对应的任务项删除 */
		mysql.DeleteTopology(db, types.DBname, taskOwner)
	}

	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(taskOwner); err != nil {
		panic(err)
	}
}

func HttpResponse(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

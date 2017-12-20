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
		panic(err)
	}

	rules, err := mysql.Select_Rules(db, types.DBname)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警规则 */
func ListUserRules(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 3) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleowner}")
		return
	}

	//var user types.RuleUser
	var user string

	user = paths[len(paths) - 1]
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	rules, err := mysql.Select_UserRules(db, types.DBname, user)
	if err != nil {
		panic(err)
	}

	if (rules == nil) {
		fmt.Println("No matched rules for owner [" + user +"] " + "or owner is non-existent")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

/* 创建新的报警规则 */
func CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule types.Rule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	/*
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
	*/
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)

	mysql.CreteTable_Rule(db, types.DBname)
	mysql.Insert_Rule(db, types.DBname,rule)
}

/* 更新指定用户指定规则名称的报警规则 */
func UpdateRule(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleowner}/{rulename}")
		return
	}

	var rule types.Rule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusResetContent)

	rule.RuleOwner = paths[len(paths) - 2]
	rule.RuleName = paths[len(paths) - 1]
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.Update_Rule(db, types.DBname, rule)
}

/* 删除指定用户指定规则名称的报警规则 */
func DeleteRule(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /rules/{ruleowner}/{rulename}")
		return
	}

	var drule types.RuleUser
	drule.RuleOwner = paths[len(paths) - 2]
	drule.RuleName = paths[len(paths) - 1]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.DeleteRule(db, types.DBname, drule)
}

/* 查看所有报警任务 */
func ListTopologys(w http.ResponseWriter, r *http.Request) {

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	topologys, err := mysql.Select_Topologys(db, types.DBname)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(topologys); err != nil {
		panic(err)
	}
}

/* 查看指定用户的所有报警任务 */
func ListUserTopologys(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 3) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyowner}")
		return
	}

	var user string
	user =  paths[len(paths) - 1]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	topologys, err := mysql.Select_UserTopologys(db, types.DBname, user)
	if err != nil {
		panic(err)
	}

	if (topologys == nil) {
		fmt.Println("No matched topologys for owner [" + user +"] " + "or owner is non-existent")
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(topologys); err != nil {
		panic(err)
	}
}

/* 创建新的报警任务 */
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
		w.WriteHeader(422) // unprocessable entity
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

/* 更新指定用户指定任务名称的报警任务 */
func UpdateTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyowner}/{topologyname}")
		return
	}

	var topology types.Topology
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &topology); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusResetContent)
	/*
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
	*/

	topology.TopologyName = paths[len(paths) - 1]
	topology.TopologyOwner = paths[len(paths) - 2]
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.Update_Topology(db, types.DBname, topology)

	//TODO：此处杀掉正在运行的任务，重新提交任务

}

/* 删除指定用户指定报警任务名称的报警任务 */
func DeleteTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 4) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyowner}/{topologyname}")
		return
	}

	var dtopology types.TopologyUser

	dtopology.TopologyOwner = paths[len(paths) - 2]
	dtopology.TopologyName = paths[len(paths) - 1]

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	mysql.DeleteTopology(db, types.DBname, dtopology)
	//TODO：此处杀掉报警任务
}


/* 根据任务名称提交报警任务 */
func SubmitTopology(w http.ResponseWriter, r *http.Request) {

	paths := strings.Split(r.URL.Path, "/")
	if (len(paths) != 5) {
		fmt.Println("URI.Path is invalid. Expected is /topologys/{topologyowner}/{topologyname}/submit")
		return
	}
	var taskOwner types.TopologyUser
	taskOwner.TopologyOwner = paths[len(paths) - 3]
	taskOwner.TopologyName = paths[len(paths) - 2]

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		panic(err)
	}

	topologys, err :=mysql.Select_SubmitTopology(db, types.DBname, taskOwner)
	if err != nil {
		fmt.Println("Error to read topology task from topology table")
		return
	}
	if topologys == nil {
		fmt.Println("No matched topology task [topologyowner=" + taskOwner.TopologyOwner + " topologyname=" +taskOwner.TopologyName + "] in topology table")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(taskOwner); err != nil {
		panic(err)
	}

	for i:=0;i<(len(topologys));i++ {

		para1 := topologys[i].TopologyName
		para2 := topologys[i].AppName
		para3 := topologys[i].KeywordIndex
		para4 := topologys[i].KeyWord
		para5 := topologys[i].TimeWindow
		para6 := topologys[i].ThresholdNum
		para7 := topologys[i].EmailList

		StringCMD := "/opt/strom/bin jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + para1 + " " + para2 + " " + strconv.Itoa(para3) + " " + para4 + " " + strconv.Itoa(para5) + " " + strconv.Itoa(para6) + " " + para7
		fmt.Println("StringCMD is: " + StringCMD)
		StringCMDTest := "echo test submit task"
		task.SubmitTask(StringCMDTest)
	}
}
package handler

import (
	"fmt"
	"net/http"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/types"
	"encoding/json"
	"io/ioutil"
	"io"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/mysql"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/redis"
	"strings"
	"strconv"
	"github.com/golang/glog"
	"time"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/utils"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/logtracer"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/pkg/DHTracer/tracing"
)

//var Tracer opentracing.Tracer

/* 服务状态检查 */
func StatusCheck(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_StatusCheck", "Enter StatusCheck", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Curl StatusCheck API to check the Status of Logalarm Server. TracerID: " + TracerId)
	fmt.Fprintln(w, "Welcome to Logalarm server!")

	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"服务正常运行"}
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}
	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave StatusCheck", false, r, Tracer)
}


/* 查看报警规则 */
func ListRules(w http.ResponseWriter, r *http.Request) {

	var Tracer *tracing.DHTracer
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_ListRules", "Enter ListRules", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Submit ListRules request to Logalarm Server. TracerID: " + TracerId)

	var requestBody types.ListRequest

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	// tracing log
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in ListRules(). TracerID: " + TracerId)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to Close request body reading failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Close request body reading failed in ListRules(). TracerID: " + TracerId)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if err := json.Unmarshal(body, &requestBody); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to unmarshal the request body failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in ListRules(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,请求数据格式错误", Result: []types.Rule{}, TotalPages: 0}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if requestBody.PageSize <= 0 || requestBody.CurrentPage <= 0 || requestBody.UserId < 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to no all necessary parameters are provide in http Request", true, r, Tracer)
		glog.Errorf("Not all necessary valid fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)", Result: []types.Rule{}, TotalPages: 0}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to connect to Mysql failed in ListRules()", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in ListRules(). TracerID: " + TracerId)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	result, count := mysql.GetCount(db, types.DBname, types.RuleTable, requestBody)
	if !result {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to get the count in table rules failed", true, r, Tracer)
		glog.Errorf("Get the count of table %s failed.TracerID: " + TracerId + " \n", types.RuleTable)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if count == 0 {
		glog.Infof("No record found in table %s. TracerID: " + TracerId + " \n", types.RuleTable)
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules with no data in table rules", false, r, Tracer)
		ret := types.HttpRespR{HttpCode: http.StatusOK, Message:"无报警规则数据", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	rules, err := mysql.SelectRules(db, types.DBname, types.RuleTable, requestBody)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules due to select data in table failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select data for all users in database [%s] failed in ListRules(). TracerID: " + TracerId, types.DBname)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.Rule{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Info("Successful processing the ListRules request to Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRespR{HttpCode: http.StatusOK, Message:"获取所有报警规则成功", Result: rules, TotalPages: ((count-1)/requestBody.PageSize + 1)}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}
	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListRules", "Leave ListRules with query result", false, r, Tracer)
}

/* 查看指定规则ID的报警规则 */
func ListSingleRule(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_ListSingleRule", "Enter ListSingleRule", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Submit ListSingleRule  request to Logalarm Server. TracerID: " + TracerId)


	var rule types.RuleId
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in ListSingleRule(). TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: nil}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in ListSingleRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result: nil}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule due to Unmarshal http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in ListSingleRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRes{HttpCode:  http.StatusInternalServerError, Message:"解析请求体失败", Result: nil}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if rule.Id <= 0  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule due to no all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)", Result: nil}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}


	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in ListSingleRule(). TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败", Result: nil}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	retRule, err := mysql.SelectRule(db, types.DBname, types.RuleTable, rule.Id)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule due to query from table rules failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select rule for [ruleid=%d] in table rules failed in ListSingleRule(). TracerID: " + TracerId, rule.Id)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: nil}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if retRule.Id == 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule with no rules found in table rules", false, r, Tracer)
		glog.Info("No rules for [ruleid=%d] in table rules found in ListSingleRule(). TracerID: " + TracerId, rule.Id)
		ret := types.HttpRes{HttpCode: http.StatusOK, Message:"无报警规则数据", Result: nil}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Infof("Successful processing the ListSingleRule [ruleId= %d] request to Logalarm Server", rule.Id)
	ret := types.HttpR{HttpCode: http.StatusOK, Message:"获取指定ID的报警规则成功", Result: retRule}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}
	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListSingleRule", "Leave ListSingleRule with query result", false, r, Tracer)
}

/* 创建新的报警规则 */
func CreateRule(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_CreateRule", "Enter CreateRule", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Submit createrule request to Logalarm Server. TracerID: " + TracerId)
	var rule types.Rule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in CreateRule(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in CreateRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Unmarshal the http request body failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in CreateRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"请求数据格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if rule.UserId <= 0 || rule.UserName == "" || rule.RuleName == "" || rule.KeyWord == "" || rule.KeywordIndex < 0  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to no all necessary parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in CreateRule(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	tx, _ := db.Begin()
	defer tx.Commit()

	ret, err := mysql.CreteDatabase(db, types.DBname); if !ret {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Create database logalarm failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Create database %s failed in CreateRule(). TracerID: " + TracerId, types.DBname)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if ! mysql.CreteTable(db , types.RuleTable, types.RuleTableCreateCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Create table rules failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Create table rules in database %s failed in CreateRule(). TracerID: " + TracerId, types.DBname)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警规则表失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	sqlCMD := types.RuleDataExistedCMD + "userid=" + strconv.Itoa(rule.UserId) + " and rulename='"+ rule.RuleName+"' limit 1"
	if mysql.RecordExisted(db, types.RuleTable, sqlCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to rule have existed", true, r, Tracer)
		glog.Errorf("Existed record for [userid=%d rulename=%s] in table %s, refuse to insert new one, try update it.  TracerID: " + TracerId + " \n", rule.UserId, rule.RuleName, types.RuleTable)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,已存在同名规则"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if ! mysql.InsertTableData(db, tx,  types.RuleTable, rule) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule due to Insert data into table rules failed", true, r, Tracer)
		glog.Errorf("Insert data into table %s [username=%s, rulename=%s] in database %s failed in CreateRule() . TracerID: " + TracerId, types.RuleTable, rule.UserName, rule.RuleName, types.DBname)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,报警规则表插入数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Info("Successful processing the CreateRule request to Logalarm Server")
	HttpResponse(w, http.StatusOK)
	retv := types.HttpRe{HttpCode: http.StatusOK, Message:"创建报警规则成功"}
	if err := json.NewEncoder(w).Encode(retv); err != nil {
		return
	}

	// 增加用户操作日志
	userLog := "创建报警规则" + rule.RuleName
	utils.RecordUserOperations(rule.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_CreateRule", "Leave CreateRule", false, r, Tracer)
}

/* 更新指定ID的报警规则 */
func UpdateRule(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_UpdateRule", "Enter UpdateRule", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	var rule types.RuleUpdate
	glog.Info("Submit UpdateRule request to Logalarm Server. TracerID: " + TracerId)

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in UpdateRule(). TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in UpdateRule(). TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in UpdateRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"请求格式错误", Result: []types.Rule{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if rule.KeyWord == "" || rule.Id <= 0  || rule.KeywordIndex < 0  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to not provide rule Keyword and rule KeywordIndex", true, r, Tracer)
		glog.Error("Need to provide rule Keyword and rule KeywordIndex. TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in UpdateRule(). TracerID: " + TracerId)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 如果规则不存在，不允许更新
	retRule, err := mysql.SelectRule(db, types.DBname, types.RuleTable, rule.Id)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to Search rule in table rules failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Search rule for [ruleid=%d] in table rules failed in UpdateRule(). TracerID: " + TracerId, rule.Id)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询目标报警规则是否存在失败", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if retRule.Id != rule.Id {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to No matched rule", true, r, Tracer)
		glog.Infof("No matched rule [ruleId= %d] in rules in UpdateRule(). TracerID: " + TracerId, rule.Id)
		ret := types.HttpRes{HttpCode: http.StatusNotFound, Message:"指定的报警规则不存在", Result: []types.Rule{}}
		HttpResponse(w, http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}


	retRule.UpdateTime = time.Now().Unix() *1000
	retRule.KeywordIndex = rule.KeywordIndex
	retRule.KeyWord = rule.KeyWord
	if !mysql.Update_Rule(db, types.DBname, retRule) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule due to Update rule in table rules failed", true, r, Tracer)
		glog.Errorf("Update rule [ruleid=%d] in table rules failed. TracerID: " + TracerId, rule.Id)
		ret := types.HttpRes{HttpCode: http.StatusInternalServerError, Message:"内部错误,更新报警规则失败", Result: []types.Rule{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	//rule.UpdateTime=rule.CreateTime
	var result = make([]types.Rule, 0)
	glog.Info("Successful processing the UpdateRule request to Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRes{HttpCode: http.StatusOK, Message:"更新报警规则成功", Result: append(result,retRule)}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}

	// 增加用户操作日志
	userLog := "更新报警规则" + retRule.RuleName
	utils.RecordUserOperations(retRule.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateRule", "Leave UpdateRule", false, r, Tracer)
}

/* 删除指定用户指定规则名称的报警规则 */
func DeleteRule(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_DeleteRule", "Enter DeleteRule", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Submit DeleteRule request to Logalarm Server. TracerID: " + TracerId)

	var rule types.RuleId

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in DeleteRule(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in DeleteRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &rule); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in DeleteRule(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if rule.Id <= 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to RuleId not provide", true, r, Tracer)
		glog.Errorf("RuleId not provide. TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供必要的字段,或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in DeleteRule(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	retRule, err := mysql.SelectRule(db, types.DBname, types.RuleTable, rule.Id)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to search rule in table rules failed", true, r, Tracer)
		glog.Error(err)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询目标报警规则失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	//if !mysql.DeleteRule(db, types.DBname, ruleId) {
	if !mysql.DeleteByID(db, types.DBname, types.RuleTable, rule.Id){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule due to Delete User rule in table rules failed", true, r, Tracer)
		glog.Errorf("Delete User rule [ruleId=%s] failed. TracerID: " + TracerId, rule.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,删除指定ID报警规则失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}else {
		glog.Info("Successful processing the DeleteRule request to Logalarm Server")
		ret := types.HttpRe{HttpCode: http.StatusOK, Message:"删除指定ID报警规则成功"}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
	}

	// 增加用户操作日志
	userLog := "删除报警规则" + retRule.RuleName
	utils.RecordUserOperations(retRule.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteRule", "Leave DeleteRule", false, r, Tracer)
}

/* 查看报警任务,如果参数中带userid则是查询用户的报警任务，如果userid字段为空，则是查整个平台的报警任务 */
func ListTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_ListTasks", "Enter ListTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit ListTasks request to Logalarm Server. TracerID: " + TracerId)
	var ReturnResult []types.RetUser
	var RetResult []types.ServiceTask
	var temp types.ServiceTask
	var requestBody types.ListRequest
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in ListTasks(). TracerID: " + TracerId)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: []types.ServiceTask{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in ListTasks(). TracerID: " + TracerId)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result:  []types.ServiceTask{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &requestBody); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Unmarshal the requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in ListTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRespT{HttpCode: http.StatusBadRequest, Message:"请求格式错误", Result:  []types.ServiceTask{}, TotalPages: 0}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if requestBody.PageSize <= 0 || requestBody.CurrentPage <= 0 || requestBody.UserId < 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Not all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary valid fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRespR{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)", Result: []types.Rule{}, TotalPages: 0}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in ListTask(). TracerID: " + TracerId)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败", Result:  []types.ServiceTask{}, TotalPages: 0}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var count int
	// 1. 获取返回的 User 信息
	if requestBody.UserId != 0 {
		//ReturnResult = append(ReturnResult,types.RetUser{ID:requestBody.UserId,})
		ReturnResult = append(ReturnResult,types.RetUser{ID:requestBody.UserId})
	}else {
		users,err := mysql.SelectUsers(db, types.DBname, types.UserTable)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select user data from table users failed", true, r, Tracer)
			glog.Error(err)
			glog.Errorf("Select user data from table %s failed in ListTasks(). TracerID: " + TracerId, types.UserTable)
			ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result:  []types.ServiceTask{}, TotalPages: 0}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		ReturnResult=append(ReturnResult,users...)
	}

	// 2. 获取 返回的 Rooms 信息
	for ku, user := range ReturnResult {
		userId := user.ID
		rooms, err := mysql.SelectRooms(db, types.DBname, types.RoomTable,userId)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select room data from table rooms failed", true, r, Tracer)
			glog.Error(err)
			glog.Errorf("Select room data from table %s failed in ListTasks(). TracerID: " + TracerId, types.RoomTable)
			ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result:  []types.ServiceTask{}, TotalPages: 0}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		ReturnResult[ku].Rooms = rooms

		// 3. 获取 Services 信息
		for kr, room := range ReturnResult[ku].Rooms {
			services, err := mysql.SelectServices(db, types.DBname, types.ServiceTable, room.ID)
			if err != nil {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select room data from table services failed", true, r, Tracer)
				glog.Error(err)
				glog.Errorf("Select room data from table %s failed in ListTasks(). TracerID: " + TracerId, types.ServiceTable)
				ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result:  []types.ServiceTask{}, TotalPages: 0}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
			ReturnResult[ku].Rooms[kr].Services = services

			count = count+len(services)

			// 4. 获取Continers 信息
			//for ks:=offset;ks<condition;ks++ {

			for ks, service := range ReturnResult[ku].Rooms[kr].Services {
				//service :=  ReturnResult[ku].Rooms[kr].Services[ks]
				temp.UserId = user.ID
				temp.UserName = user.UserName
				temp.RoomId = room.ID
				temp.RoomName = room.RoomName
				temp.ServiceName = service.ServiceName
				temp.ServiceId = service.ID
				temp.TaskNameByUser = service.TaskNameByUser

				containers, err :=  mysql.SelectContainers(db, types.DBname, types.ContainerTable, service.ID)
				if err != nil {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select container data from table containers failed", true, r, Tracer)
					glog.Error(err)
					glog.Errorf("Select container data from table %s failed in ListTasks(). TracerID: " + TracerId, types.ContainerTable)
					ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result:  []types.ServiceTask{}, TotalPages: 0}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
				ReturnResult[ku].Rooms[kr].Services[ks].Containers = containers

				// 5. 获取 Files 信息

				for kc, container := range ReturnResult[ku].Rooms[kr].Services[ks].Containers{
					files, err :=  mysql.SelectFiles(db, types.DBname, types.FileTable, container.ID)
					if err != nil {
						Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select file data from table files failed", true, r, Tracer)
						glog.Error(err)
						glog.Errorf("Select file data from table %s failed in ListTasks(). TracerID: " + TracerId, types.FileTable)
						ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result:  []types.ServiceTask{}, TotalPages: 0}
						HttpResponse(w, http.StatusInternalServerError)
						if err := json.NewEncoder(w).Encode(ret); err != nil {
							return
						}
						return
					}
					if len(files) != 0 {
						ReturnResult[ku].Rooms[kr].Services[ks].Containers[kc].Files = files
					}else{
						ReturnResult[ku].Rooms[kr].Services[ks].Containers[kc].Files = []types.RetFile{}
					}
					// 6. 获取 Tasks 信息
					for kf, file := range ReturnResult[ku].Rooms[kr].Services[ks].Containers[kc].Files {
						tasks, err := mysql.SelectFileTasks(db, types.DBname, types.TaskTable, file.Id)
						if err != nil {
							Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to Select file data from table tasks failed", true, r, Tracer)
							glog.Error(err)
							glog.Errorf("Select file data from table %s failed in ListTasks(). TracerID: " + TracerId, types.TaskTable)
							ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}, TotalPages: 0}
							HttpResponse(w, http.StatusInternalServerError)
							if err := json.NewEncoder(w).Encode(ret); err != nil {
								return
							}
							return
						}
						// 重新将返回结果中时间由秒换算成分钟
						for kt := range tasks {
							tasks[kt].TimeWindow = tasks[kt].TimeWindow / 60
						}

						if len(tasks) !=0 {
							ReturnResult[ku].Rooms[kr].Services[ks].Containers[kc].Files[kf].Tasks = tasks
						} else {
							ReturnResult[ku].Rooms[kr].Services[ks].Containers[kc].Files[kf].Tasks = []types.Task{}
						}
						//count = count + len(tasks)
					}

				}

				temp.Containers = ReturnResult[ku].Rooms[kr].Services[ks].Containers
				RetResult = append(RetResult,temp)
			}

		}

	}

	if len(RetResult) == 0{
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks due to no related data in table tasks", false, r, Tracer)
		ret := types.HttpRespT{HttpCode: http.StatusOK, Message:"无相关报警任务", Result: []types.ServiceTask{}, TotalPages: 0}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	offset := (requestBody.CurrentPage - 1) * requestBody.PageSize
	var end int
	if offset + requestBody.PageSize <= count {
		end = offset + requestBody.PageSize
	} else {
		end = count
	}
	Res := RetResult[offset:end]

	glog.Info("Successful processing the ListTasks request to Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRespT{HttpCode: http.StatusOK, Message:"获取所有报警任务成功", Result: Res, TotalPages:  ((count-1)/requestBody.PageSize + 1)}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListTasks", "Leave ListTasks", false, r, Tracer)
}

/* 以服务为单位提交报警任务 */
func SubmitTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_SubmitTasks", "Enter SubmitTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit SubmitTasks requests to Logalarm Server. TracerID: " + TracerId)

	var submitTasks types.SubmitTasks
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed SubmitTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in SubmitTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &submitTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in SubmitTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	//参数有效性判断
	if submitTasks.UserId <=0 || submitTasks.UserName== "" || submitTasks.TaskNameByUser == ""  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供UserName、UserId、TaskName等信息,或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if submitTasks.RoomInfo.RoomName == "" || submitTasks.RoomInfo.Id  <=0 || submitTasks.RoomInfo.ServiceInfo.Id <=0 || submitTasks.RoomInfo.ServiceInfo.ServiceName == "" {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供RoomName、RoomId、ServiceID、ServiceID等信息,或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if len(submitTasks.RoomInfo.ServiceInfo.Containers) == 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供服务的Containers信息)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	for _, container := range submitTasks.RoomInfo.ServiceInfo.Containers {
		if container.ContainerName == "" || container.Id <=0 {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Invalid http request parameters", true, r, Tracer)
			HttpResponse(w, http.StatusBadRequest)
			ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供ContainerID、containerName等信息或者字段值非法)"}
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		for _,file := range container.Files {
			for _,task:= range file.Tasks {
				if task.TimeWindow  <=0 || task.ThresholdNum <=0 || task.AlarmGroupID  <=0 || task.RuleId <=0 {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Invalid http request parameters", true, r, Tracer)
					HttpResponse(w, http.StatusBadRequest)
					ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供报警配置等信息(报警周期、阈值、报警组ID、报警规则ID)或者字段值非法)"}
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
			}
 		}
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}


	// 创建 TaskName 表
	if !mysql.CreteTable(db, types.TaskNameTable, types.TaskNameTableCreateCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table tasknames failed", true, r, Tracer)
		glog.Errorf("Create table %s failed with CMD[%s]. TracerID: " + TracerId, types.TaskNameTable, types.TaskNameTableCreateCMD)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建容器云用户服务任务表失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 如果已存在同名任务，不允许创建
	var taskNames []types.TaskNameByUser
	var sqlCMD string= "select userid,taskname from "+ types.DBname +"." + types.TaskNameTable + " where userid=" + strconv.Itoa(submitTasks.UserId) + " and taskname='" + submitTasks.TaskNameByUser + "'"
	//rows, err := db.Query("select taskname from "+ types.DBname +"." + types.UserTable + "where id=" + strconv.Itoa(submitTasks.UserId))
	rows, err := db.Query(sqlCMD)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Select data from table users failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select data from %s failed in SubmitTasks(). TracerID: " + TracerId, types.UserTable)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取任务名失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	defer rows.Close()

	var taskNameByUser types.TaskNameByUser
	for rows.Next() {
		err := rows.Scan(&taskNameByUser.UserId, &taskNameByUser.TaskNameByUser)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to try to get User defined taskname failed", true, r, Tracer)
			glog.Error(err)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"获取任务名失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		taskNames = append(taskNames,taskNameByUser)
	}
	if len(taskNames) !=0 {
		for _,existedTaskName := range taskNames{
			if existedTaskName.TaskNameByUser  == submitTasks.TaskNameByUser {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to User defined taskname existed", true, r, Tracer)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"已存在同名任务"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
			return
			}
		}

	}


	// user 信息
	var user types.User
	user.Id = submitTasks.UserId
	user.UserName=submitTasks.UserName

	//room信息
	submitTasks.RoomInfo.UserId=submitTasks.UserId

	//service 信息
	submitTasks.RoomInfo.ServiceInfo.RoomId=submitTasks.RoomInfo.Id

	// 如果提交的创建日志报警任务服务中实际报警任务数量为0,则不创建任何服务或数据
	var taskCount int=0
	for _,c := range submitTasks.RoomInfo.ServiceInfo.Containers {
		for _,f := range c.Files{
			taskCount = taskCount + len(f.Tasks)
		}
	}
	if taskCount == 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to No tasks need to submit", true, r, Tracer)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"无报警任务需要创建"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	tx, _ := db.Begin()
	defer tx.Commit()
	for _, con := range submitTasks.RoomInfo.ServiceInfo.Containers{
		con.ServiceId=submitTasks.RoomInfo.ServiceInfo.Id

		for  _, file := range con.Files {

			file.ContainerId=con.Id
			file.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id

			if !mysql.CreteTable(db, types.FileTable, types.FileTableCreateCMD) {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table files failed", true, r, Tracer)
				glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.FileTable, types.FileTableCreateCMD)
				//删除相关的任务和数据
				mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警文件表失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}


			tx1, _ := db.Begin()
			if ! mysql.InsertTableData(db, tx1, types.FileTable, file){
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table files failed", true, r, Tracer)
				glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.FileTable)
				//删除相关的任务和数据
				mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警文件表数据失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
			tx1.Commit()

			// 获取文件在File表中对应的ID
			//rows,_ := db.Query("select id from " + types.DBname + "." + types.FileTable + " where filename='" + file.FileName + "' and filepath='" +file.FilePath + "' and cid=" + file.ContainerId + " and sid=" +file.ServiceId )
			var fid int
			//err = db.QueryRow("select name from user where id = ?", 1).Scan(&name)
			//sqlCMD := "select id from " + types.DBname + "." + types.FileTable + " where filename='" + file.FileName + "' and filepath='" +file.FilePath + "' and cid=" + file.ContainerId + " and sid=" +file.ServiceId
			//fmt.Printf("*************FileName is: %s \n", file.FileName)
			//fmt.Printf("*************FilePath is: %s \n", file.FilePath)
			//fmt.Printf("*************ContainerID is: %d \n", file.ContainerId)
			//fmt.Printf("*************ServiceID is: %d \n", file.ServiceId)
			sqlCMD := "select id from " + types.DBname + "." + types.FileTable + " where filename=? and filepath=? and cid=? and sid=?"
			//fmt.Printf("***********FILEID GET sqlCMD is: %s \n", sqlCMD)
			err = db.QueryRow(sqlCMD, file.FileName,file.FilePath, file.ContainerId,file.ServiceId).Scan(&fid)
			if err !=nil {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Get the fileID failed", true, r, Tracer)
				glog.Error(err)
				glog.Error("Get the fileID failed in SubmitTasks(). TracerID: " + TracerId)
				mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取文件ID数据失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}


			for _, Ttask := range file.Tasks {
				rule, err := mysql.SelectRule(db, types.DBname, types.RuleTable, Ttask.RuleId)
				if err != nil {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Get rule in table rules failed", true, r, Tracer)
					glog.Error(err)
					glog.Errorf("Select rule for [ruleid=%d] in table rules failed in SubmitTasks(). TracerID: " + TracerId, Ttask.RuleId)
					//删除相关的任务和数据
					mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询报警规则数据失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}

				Ttask.KeyWord = rule.KeyWord
				Ttask.KeywordIndex = rule.KeywordIndex
				Ttask.FileName = file.FileName
				Ttask.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id
				Ttask.ContainerId = con.Id
				Ttask.FileId = fid
				Ttask.RoomId = submitTasks.RoomInfo.Id
				Ttask.UserId = submitTasks.UserId
				Ttask.FilePath = file.FilePath
				Ttask.TimeWindow = Ttask.TimeWindow * 60
				Ttask.TaskState = "active"

				var topicName string
				var taskName string
				if Ttask.FileName != "" {
					topicName = submitTasks.UserName + "-" + strconv.Itoa(submitTasks.UserId) + "." + submitTasks.RoomInfo.ServiceInfo.ServiceName + "." + con.ContainerName + "." + file.FileName
					taskName = strconv.Itoa(Ttask.UserId)+ "-" + strconv.Itoa(Ttask.ServiceId) + "-" + strconv.Itoa(Ttask.ContainerId) + "-" + strconv.Itoa(fid)+ "-" + strconv.Itoa(rule.Id)+ "-"+ strconv.Itoa(Ttask.AlarmGroupID) + "-" + strconv.Itoa(Ttask.TimeWindow) + "-" + strconv.Itoa(Ttask.ThresholdNum)
				}else {
					topicName = submitTasks.UserName + "-" + strconv.Itoa(submitTasks.UserId) + "." + submitTasks.RoomInfo.ServiceInfo.ServiceName + "." + con.ContainerName
					taskName = strconv.Itoa(Ttask.UserId)+ "-" + strconv.Itoa(Ttask.ServiceId) + "-" + strconv.Itoa(Ttask.ContainerId) + "-" + strconv.Itoa(rule.Id)+ "-" + strconv.Itoa(Ttask.AlarmGroupID) + "-" + strconv.Itoa(Ttask.TimeWindow) + "-" + strconv.Itoa(Ttask.ThresholdNum)
				}

				Ttask.TaskName=taskName + "-" + Ttask.KeyWord


				// 避免出现提交任务名不相同，但服务底层实际任务相同的任务
				rows, err = db.Query("select * from " + types.DBname + "." + types.TaskTable + " where taskname='" + Ttask.TaskName + "'")
				if err != nil {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Check task existed in table tasks failed", true, r, Tracer)
					glog.Error(err)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询是否有相同配置报警任务失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
				existedTasks := mysql.Return_Tasks(rows)
				rows.Close()
				if len(existedTasks) != 0{
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to task have existed in table tasks", true, r, Tracer)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,该服务已有相同配置报警任务"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}

				contacts, err :=redis.GetPhoneAndEmailInfo(Ttask.AlarmGroupID); if err !=nil{
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to get Phone and Email info failed", true, r, Tracer)
					glog.Error(err)
					//删除相关的任务和数据
					mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取报警任务组信息失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
				//contacts := []types.Contact{{"123456789", "123@.456.com"},{"1234567","234@567.com"}}
				var emails  []string
				for _, contact := range contacts{
					emails = append(emails, contact.Mail)
				}
				emailList := strings.Join(emails, ",")

				StringCMD := "/opt/storm/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + Ttask.TaskName + " " + topicName + " " + strconv.Itoa(rule.KeywordIndex) +
					" " + rule.KeyWord + " " + strconv.Itoa(Ttask.TimeWindow) + " " + strconv.Itoa(Ttask.ThresholdNum) + " " + emailList
				glog.Infof("提交任务命令： %s. TracerID: " + TracerId, StringCMD)
				if !utils.TopologySubmit(StringCMD, taskName) {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Submit Topology task failed", true, r, Tracer)
					glog.Errorf("Submit Topology task %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, taskName, StringCMD)
					//删除相关的任务和数据
					mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,提交报警任务失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
				if !mysql.CreteTable(db, types.TaskTable, types.TaskTableCreateCMD){
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table tasks failed", true, r, Tracer)
					glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.TaskTable, types.TaskTableCreateCMD)
					//删除相关的任务和数据
					mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警任务表失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
				/*
				Ttask.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id
				Ttask.ContainerId = con.Id
				Ttask.FileId = fid
				Ttask.RoomId = submitTasks.RoomInfo.Id
				Ttask.UserId = submitTasks.UserId
				Ttask.FilePath = file.FilePath
				*/
				if ! mysql.InsertTableData(db, tx, types.TaskTable, Ttask){
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table tasks failed", true, r, Tracer)
					glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.TaskTable)
					//删除相关的任务和数据
					mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警任务表数据失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
			}
		}

		if !mysql.CreteTable(db, types.ContainerTable, types.ContainerTableCreateCMD) {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table containers failed", true, r, Tracer)
			glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.ContainerTable, types.ContainerTableCreateCMD)
			//删除相关的任务和数据
			mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警任务容器表失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		if !mysql.InsertTableData(db, tx, types.ContainerTable, con){
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table containers failed", true, r, Tracer)
			glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.ContainerTable)
			//删除相关的任务和数据
			mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警任务容器表数据失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
	}

	if !mysql.CreteTable(db, types.ServiceTable, types.ServiceTableCreateCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table services failed", true, r, Tracer)
		glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.ServiceTable, types.ServiceTableCreateCMD)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建容器云应用服务表失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	submitTasks.RoomInfo.ServiceInfo.TaskNameByUser = submitTasks.TaskNameByUser
	if !mysql.InsertTableData(db, tx, types.ServiceTable, submitTasks.RoomInfo.ServiceInfo){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table services failed", true, r, Tracer)
		glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.ServiceTable)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入容器云应用服务表数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 创建 Room 表
	if !mysql.CreteTable(db, types.RoomTable, types.RoomTableCreateCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table rooms failed", true, r, Tracer)
		glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.RoomTable, types.RoomTableCreateCMD)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建容器云应用Room表失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if ! mysql.InsertTableData(db, tx, types.RoomTable, submitTasks.RoomInfo){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table rooms failed", true, r, Tracer)
		glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.ServiceTable)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入容器云应用Room表数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 创建 User 表
	if !mysql.CreteTable(db, types.UserTable, types.UserTableCreateCMD) {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Create table users failed", true, r, Tracer)
		glog.Errorf("Create table %s failed with CMD[%s] in SubmitTasks(). TracerID: " + TracerId, types.UserTable, types.UserTableCreateCMD)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建User表失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if ! mysql.InsertTableData(db, tx, types.UserTable,user){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table users failed", true, r, Tracer)
		glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.ServiceTable)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入容器云应用User表数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var taskNameData types.TaskNameByUser
	taskNameData.TaskNameByUser = submitTasks.TaskNameByUser
	taskNameData.UserId = submitTasks.UserId
	taskNameData.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id
	if ! mysql.InsertTableData(db, tx, types.TaskNameTable,taskNameData){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks due to Insert data into table tasknames failed", true, r, Tracer)
		glog.Errorf("Insert data into table %s failed in SubmitTasks(). TracerID: " + TracerId, types.TaskNameTable)
		//删除相关的任务和数据
		mysql.DeleteServiceRalatedData(db, types.DBname, submitTasks.RoomInfo.ServiceInfo.Id)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入TaskName表数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Info("Successful processing the SubmitTopology request to Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"服务的所有日志报警任务创建成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}

	// 增加用户操作日志
	userLog := "提交报警任务" + submitTasks.TaskNameByUser
	utils.RecordUserOperations(submitTasks.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_SubmitTasks", "Leave SubmitTasks", false, r, Tracer)
}


/* 以服务为单位提更新报警任务 */
func UpdateServiceTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_UpdateServiceTasks", "Enter UpdateServiceTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit UpdateServiceTasks requests to Logalarm Server. TracerID: " + TracerId)

	var submitTasks types.SubmitTasks
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in UpdateServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in UpdateServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &submitTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in UpdateServiceTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	//参数有效性判断
	if submitTasks.UserId <=0 || submitTasks.UserName== "" || submitTasks.TaskNameByUser == ""  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供UserName、UserId、TaskName等信息,或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if submitTasks.RoomInfo.RoomName == "" || submitTasks.RoomInfo.Id  <=0 || submitTasks.RoomInfo.ServiceInfo.Id <=0 || submitTasks.RoomInfo.ServiceInfo.ServiceName == "" {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供RoomName、RoomId、ServiceID、ServiceID等信息,或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if len(submitTasks.RoomInfo.ServiceInfo.Containers) == 0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Invalid http request parameters", true, r, Tracer)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供服务的Containers信息)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	for _, container := range submitTasks.RoomInfo.ServiceInfo.Containers {
		if container.ContainerName == "" || container.Id <=0 {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Invalid http request parameters", true, r, Tracer)
			HttpResponse(w, http.StatusBadRequest)
			ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供ContainerID、containerName等信息或者字段值非法)"}
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		for _,file := range container.Files {
			for _,task:= range file.Tasks {
				if task.TimeWindow  <=0 || task.ThresholdNum <=0 || task.AlarmGroupID  <=0 || task.RuleId <=0 {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Invalid http request parameters", true, r, Tracer)
					HttpResponse(w, http.StatusBadRequest)
					ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误(未提供报警配置等信息(报警周期、阈值、报警组ID、报警规则ID)或者字段值非法)"}
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
			}
		}
	}

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in UpdateServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 如果指定的服务对应的用户提交任务名不存在，不允许更新
	var taskNames []types.TaskNameByUser
	var sqlCMD string= "select userid,taskname from "+ types.DBname +"." + types.TaskNameTable + " where userid=" + strconv.Itoa(submitTasks.UserId) + " and taskname='" + submitTasks.TaskNameByUser + "'"
	//rows, err := db.Query("select taskname from "+ types.DBname +"." + types.UserTable + "where id=" + strconv.Itoa(submitTasks.UserId))
	rows, err := db.Query(sqlCMD)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Select data from table users failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select data from %s failed. TracerID: " + TracerId, types.UserTable)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取任务名失败"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	defer rows.Close()

	var taskNameByUser types.TaskNameByUser
	for rows.Next() {
		err := rows.Scan(&taskNameByUser.UserId, &taskNameByUser.TaskNameByUser)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to get User derfined taskname failed", true, r, Tracer)
			glog.Error(err)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取任务名失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		taskNames = append(taskNames,taskNameByUser)
	}
	if len(taskNames) ==0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to User derfined taskname not existed", true, r, Tracer)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"对应的任务不存在"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 获取服务当前已存在的任务列表
	rows, err = db.Query("select * from " + types.DBname + "." + types.TaskTable + " where serviceid=" + strconv.Itoa(submitTasks.RoomInfo.ServiceInfo.Id))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Select data from table tasks failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select data from %s failed in UpdateServiceTasks(). TracerID: " + TracerId, types.TaskTable)
		return
	}
	defer rows.Close()
	existedTasks := mysql.Return_Tasks(rows)

	// 记录更新请求处理过程中需要保留的Task
	var toKeepTasks = make([]types.Task, 0)

	// user 信息
	var user types.User
	user.Id = submitTasks.UserId
	user.UserName=submitTasks.UserName

	//room信息
	submitTasks.RoomInfo.UserId=submitTasks.UserId

	//service 信息
	submitTasks.RoomInfo.ServiceInfo.RoomId=submitTasks.RoomInfo.Id

	// 记录需要保留的文件，
	var filesToAdd = make([]int,0)

	// 记录当前已经在的文件
	var filesExited = make([]int,0 )
	// 获取当前已有任务关联的文件的列表
	for _, ex := range existedTasks {
		filesExited = append(filesExited, ex.FileId)
	}


	tx, _ := db.Begin()
	defer tx.Commit()
	for _, con := range submitTasks.RoomInfo.ServiceInfo.Containers{
		con.ServiceId=submitTasks.RoomInfo.ServiceInfo.Id

		for  _, file := range con.Files {
			file.ContainerId=con.Id
			file.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id

			if !mysql.CreteTable(db, types.FileTable, types.FileTableCreateCMD) {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Create table files failed", true, r, Tracer)
				glog.Errorf("Create table %s failed with CMD[%s] in UpdateServiceTasks(). TracerID: " + TracerId, types.FileTable, types.FileTableCreateCMD)
				//删除相关的任务和数据
				//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警文件表失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}


			tx1, _ := db.Begin()
			if ! mysql.InsertTableData(db, tx1, types.FileTable, file){
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Insert data into table files failed", true, r, Tracer)
				glog.Errorf("Insert data into table %s failed in UpdateServiceTasks(). TracerID: " + TracerId, types.FileTable)
				//删除相关的任务和数据
				//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警文件表数据失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
			tx1.Commit()


			// 获取文件在File表中对应的ID
			var fid int
			sqlCMD := "select id from " + types.DBname + "." + types.FileTable + " where filename=? and filepath=? and cid=? and sid=?"
			err = db.QueryRow(sqlCMD, file.FileName,file.FilePath, file.ContainerId,file.ServiceId).Scan(&fid)
			if err !=nil {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Get the fileID in talbe files failed", true, r, Tracer)
				glog.Error(err)
				glog.Error("Get the fileID failed in UpdateServiceTasks(). TracerID: " + TracerId)
				//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误，获取报警文件ID数据失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}

			// 记录需要保留的文件
			filesToAdd=append(filesToAdd,fid)


			for _, Ttask := range file.Tasks {
				rule, err := mysql.SelectRule(db, types.DBname, types.RuleTable, Ttask.RuleId)
				if err != nil {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Select rule in table rules failed", true, r, Tracer)
					glog.Error(err)
					glog.Errorf("Select rule for [ruleid=%d] in table rules failed in UpdateServiceTasks(). TracerID: " + TracerId, Ttask.RuleId)
					//删除相关的任务和数据
					//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询报警规则数据失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}

				Ttask.KeyWord = rule.KeyWord
				Ttask.KeywordIndex = rule.KeywordIndex
				Ttask.FileName = file.FileName
				Ttask.ServiceId = submitTasks.RoomInfo.ServiceInfo.Id
				Ttask.ContainerId = con.Id
				Ttask.FileId = fid
				Ttask.RoomId = submitTasks.RoomInfo.Id
				Ttask.UserId = submitTasks.UserId
				Ttask.FilePath = file.FilePath
				Ttask.TimeWindow = Ttask.TimeWindow * 60
				Ttask.TaskState = "active"
				//fmt.Printf("++++++++++FileName: %s \n",Ttask.FileName)
				//fmt.Printf("++++++++++Keyword: %s \n",Ttask.KeyWord)
				//fmt.Printf("++++++++++KeywordIndex: %d \n",Ttask.KeywordIndex)
				var topicName string
				var taskName string
				if Ttask.FileName != "" {
					topicName = submitTasks.UserName + "-" + strconv.Itoa(submitTasks.UserId) + "." + submitTasks.RoomInfo.ServiceInfo.ServiceName + "." + con.ContainerName + "." + file.FileName
					taskName = strconv.Itoa(Ttask.UserId)+ "-" + strconv.Itoa(Ttask.ServiceId) + "-" + strconv.Itoa(Ttask.ContainerId) + "-" + strconv.Itoa(fid)+ "-" + strconv.Itoa(rule.Id)+ "-" + strconv.Itoa(Ttask.AlarmGroupID) + "-" + strconv.Itoa(Ttask.TimeWindow) + "-" + strconv.Itoa(Ttask.ThresholdNum)
				}else {
					topicName = submitTasks.UserName + "-" + strconv.Itoa(submitTasks.UserId) + "." + submitTasks.RoomInfo.ServiceInfo.ServiceName + "." + con.ContainerName
					taskName = strconv.Itoa(Ttask.UserId)+ "-" + strconv.Itoa(Ttask.ServiceId) + "-"+ strconv.Itoa(Ttask.ContainerId) + "-" + strconv.Itoa(rule.Id)+ "-" + strconv.Itoa(Ttask.AlarmGroupID) + "-" + strconv.Itoa(Ttask.TimeWindow) + "-" + strconv.Itoa(Ttask.ThresholdNum)
				}

				Ttask.TaskName=taskName + "-" + Ttask.KeyWord

				//fmt.Printf("++++++++++TaskName: %s \n",Ttask.TaskName)
				var needCreate bool
				for _, task := range existedTasks {
					needCreate= true
					//if task.TaskName == Ttask.TaskName && task.TimeWindow == Ttask.TimeWindow && task.ThresholdNum == Ttask.ThresholdNum && task.AlarmGroupID == Ttask.AlarmGroupID  && task.FilePath == Ttask.FilePath && task.KeyWord ==Ttask.KeyWord {
					if task.TaskName == Ttask.TaskName {
						glog.Infof("Task for [Service:%s Contariner:%s Filepath: %s Filename:%s Keyword:%s Keyeordindex:%d Interval:%d Thresholdnum:%d] is running now. No need to update. TracerID: " + TracerId,
							submitTasks.RoomInfo.ServiceInfo.ServiceName, con.ContainerName, file.FilePath, file.FileName, Ttask.KeyWord, Ttask.KeywordIndex, Ttask.TimeWindow, Ttask.ThresholdNum)
						toKeepTasks=append(toKeepTasks,task)
						needCreate= false
						break
					}
				}

				if needCreate {
					contacts, err :=redis.GetPhoneAndEmailInfo(Ttask.AlarmGroupID); if err !=nil{
						Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Get Phone and Emails info failed", true, r, Tracer)
						glog.Error(err)
						//删除相关的任务和数据
						//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
						ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,获取报警任务组信息失败"}
						HttpResponse(w, http.StatusInternalServerError)
						if err := json.NewEncoder(w).Encode(ret); err != nil {
							return
						}
						return
					}
					//contacts := []types.Contact{{"123456789", "123@.456.com"},{"1234567","234@567.com"}}
					var emails  []string
					for _, contact := range contacts{
						emails = append(emails, contact.Mail)
					}
					emailList := strings.Join(emails, ",")

					StringCMD := "/opt/storm/bin/storm jar /opt/storm/jar/logalarm.jar com.dahua.logalarm.LogAlarmTopology " + Ttask.TaskName + " " + topicName + " " + strconv.Itoa(rule.KeywordIndex) +
						" " + rule.KeyWord + " " + strconv.Itoa(Ttask.TimeWindow) + " " + strconv.Itoa(Ttask.ThresholdNum) + " " + emailList
					glog.Infof("提交任务命令： %s. TracerID: " + TracerId, StringCMD)

					if !utils.TopologySubmit(StringCMD, taskName) {
						Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Submit Topology task failed", true, r, Tracer)
						glog.Errorf("Submit Topology task %s failed with CMD[%s] in UpdateServiceTasks(). TracerID: " + TracerId, taskName, StringCMD)
						//删除相关的任务和数据
						//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
						ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,提交报警任务失败"}
						HttpResponse(w, http.StatusInternalServerError)
						if err := json.NewEncoder(w).Encode(ret); err != nil {
							return
						}
						return
					}

					if !mysql.CreteTable(db, types.TaskTable, types.TaskTableCreateCMD){
						Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Create table tasks failed", true, r, Tracer)
						glog.Errorf("Create table %s failed with CMD[%s] in UpdateServiceTasks(). TracerID: " + TracerId, types.TaskTable, types.TaskTableCreateCMD)
						//删除相关的任务和数据
						//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
						ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警任务表失败"}
						HttpResponse(w, http.StatusInternalServerError)
						if err := json.NewEncoder(w).Encode(ret); err != nil {
							return
						}
						return
					}

					if ! mysql.InsertTableData(db, tx, types.TaskTable, Ttask){
						Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Insert data into table tasks failed", true, r, Tracer)
						glog.Errorf("Insert data into table %s failed in UpdateServiceTasks(). TracerID: " + TracerId, types.TaskTable)
						//删除相关的任务和数据
						//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
						ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警任务表数据失败"}
						HttpResponse(w, http.StatusInternalServerError)
						if err := json.NewEncoder(w).Encode(ret); err != nil {
							return
						}
						return
					}
				}
			}
		}

		if !mysql.CreteTable(db, types.ContainerTable, types.ContainerTableCreateCMD) {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Create table containers failed", true, r, Tracer)
			glog.Errorf("Create table %s failed with CMD[%s] in UpdateServiceTasks(). TracerID: " + TracerId, types.ContainerTable, types.ContainerTableCreateCMD)
			//删除相关的任务和数据
			//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,创建报警任务容器表失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		if !mysql.InsertTableData(db, tx, types.ContainerTable, con){
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Insert data into table containers failed", true, r, Tracer)
			glog.Errorf("Insert data into table %s failed in UpdateServiceTasks(). TracerID: " + TracerId, types.ContainerTable)
			//删除相关的任务和数据
			//mysql.DeleteServiceRalatedData(db, types.DBname, con.ServiceId)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,插入报警任务容器表数据失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
	}

	// 删除更新过程中无需保留的任务
	if len(toKeepTasks) != len(existedTasks){
		for _, existed := range existedTasks {
			toKill :=true
			for _, keep := range toKeepTasks{
				if keep.Id == existed.Id {
					toKill = false
				}
			}

			if toKill {
				StringCMD := "/opt/storm/bin/storm kill " + existed.TaskName
				if !utils.TopologySubmit(StringCMD, existed.TaskName) {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Submit kill Topology task failed", true, r, Tracer)
					glog.Errorf("Submit kill Topology task %s failed with CMD[%s] in UpdateServiceTasks(). TracerID: " + TracerId, existed.TaskName, StringCMD)
					ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,杀死过期任务失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}

				_ ,err =tx.Exec("delete from " +  types.DBname  + "." + types.TaskTable + " where id=?", existed.Id)
				if err != nil {
					Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Delete task data in table tasks failed", true, r, Tracer)
					glog.Error(err)
					glog.Error("Delete tasks data in table failed in UpdateServiceTasks(). TracerID: " + TracerId)
					ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,删除任务表数据失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
			}

			// 删除文件表中属于此服务的无报警任务的文件
			/*
			fileToKill := true
			for _, keepT := range toKeepTasks{
				if keepT.FileId == existed.FileId {
					fileToKill = false
				}
			}

			if fileToKill {
				_ ,err =tx.Exec("delete from " +  types.DBname  + "." + types.FileTable + " where id=?", existed.FileId)
				if err != nil {
					glog.Error(err)
					glog.Error("Delete file data in table failed in UpdateServiceTasks()")
					ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,删除任务表数据失败"}
					HttpResponse(w, http.StatusInternalServerError)
					if err := json.NewEncoder(w).Encode(ret); err != nil {
						return
					}
					return
				}
			}*/

		}
	}

	// 仍然需要保留的任务对应的文件列表
	filesToKeep := make([]int,0)
	for _,kT := range toKeepTasks{
		filesToKeep = append(filesToKeep, kT.FileId)
	}

	// 如果文件在 filesExited 中，但不在filesToAdd  或 filesToKeep 中，则删除
	for _, Fid := range filesExited {
		fileToKill := true
		for _, AddFid := range filesToAdd {
			if Fid == AddFid{
				fileToKill = false
				break
			}
		}

		for _, KeepFid := range filesToAdd {
			if Fid == KeepFid{
				fileToKill = false
				break
			}
		}

		if fileToKill{
			_ ,err =tx.Exec("delete from " +  types.DBname  + "." + types.FileTable + " where id=?", Fid)
			if err != nil {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks due to Delete file data in table failed", true, r, Tracer)
				glog.Error(err)
				glog.Error("Delete file data in table failed in UpdateServiceTasks(). TracerID: " + TracerId)
				ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,删除任务表数据失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
		}

	}


	glog.Info("Successful processing the UpdateServiceTasks request to Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"服务的日志报警任务更新成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}

	// 增加用户操作日志
	userLog := "更新报警任务" + submitTasks.TaskNameByUser
	utils.RecordUserOperations(submitTasks.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_UpdateServiceTasks", "Leave UpdateServiceTasks", false, r, Tracer)
}

/* 删除指定服务ID的所有报警任务 */
func DeleteServiceTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_DeleteServiceTasks", "Enter DeleteServiceTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)
	glog.Info("Submit Delete Service's Logalarm Tasks to Logalarm Server. TracerID: " + TracerId)

	var serviceTasks types.ServiceID
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in DeleteServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in DeleteServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &serviceTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in DeleteServiceTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRe{HttpCode: http.StatusBadRequest, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if serviceTasks.ServiceId <=0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks due to Not all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary valid fileds are provide, need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	serviceId := serviceTasks.ServiceId


	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in DeleteServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 获取要删除的报警任务的用户id和对应的用户提交报警任务名称，用于用户日志信息收集
	taskNameByUser,_ := mysql.SelectTaskNameByUser(db, types.DBname, types.TaskNameTable, serviceId)

	if !mysql.DeleteServiceRalatedData(db, types.DBname, serviceId){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks due to Delete service related data and tasks failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Delete service related data and tasks failed in DeleteServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,删服务及任务相关数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Info("Successful to delete all the tasks of this service")
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"删除服务相关任务成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}
	// 增加用户操作日志
	userLog := "删除报警任务" + taskNameByUser.TaskNameByUser
	utils.RecordUserOperations(taskNameByUser.UserId, userLog)

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteServiceTasks", "Leave DeleteServiceTasks", false, r, Tracer)
	return
}


/* 查看指定服务ID下的所有报警任务 */
func ListServiceTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_ListServiceTasks", "Enter ListServiceTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Infof("Submit ListServiceTasks request to Logalarm Server. TracerID: " + TracerId)

	var serviceTasks types.ServicesInfo
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in ListServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in ListServiceTasks(). TracerID: " + TracerId)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &serviceTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in ListServiceTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusBadRequest)
		ret := types.HttpRespT{HttpCode: http.StatusBadRequest, Message:"请求格式错误", Result: []types.ServiceTask{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if serviceTasks.ServiceId <=0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Not all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary valid fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)", Result: []types.ServiceTask{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	serviceId := serviceTasks.ServiceId

	var RetResult []types.ServiceTask
	var temp types.ServiceTask

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in ListServiceTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusUnprocessableEntity)
		ret := types.HttpRespT{HttpCode: http.StatusBadRequest, Message:"连接数据库错误", Result: []types.ServiceTask{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var ReturnResult []types.RetUser

	// 1. 获取serviceName 和 roomID 信息
	var serviceName string
	var roomId int
	var taskNameByUser string // 用户定义的以服务为单位的任务名称
	err = db.QueryRow("select servicename,rid,taskname from " + types.DBname + "." + types.ServiceTable + " where id=?", serviceId).Scan(&serviceName,&roomId, &taskNameByUser)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Get ServiceName and RoomId from table services failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Get ServiceName and RoomId from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.ServiceTable)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var services []types.RetService
	service := types.RetService{ID: serviceId, ServiceName: serviceName, TaskNameByUser: taskNameByUser}
	services = append(services,service)

	// 2. 获取roomName 和 userId 信息
	var roomName string
	var userId int
	err = db.QueryRow("select roomname,uid from " + types.DBname + "." + types.RoomTable + " where id=?", roomId).Scan(&roomName,&userId)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Get RoomName and UserId from table rooms failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Get RoomName and UserId from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.RoomTable)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var rooms []types.RetRoom
	room := types.RetRoom{ID: roomId,RoomName: roomName, Services: services}
	rooms = append(rooms,room)
	// 3. 获取 userName 信息
	var userName string
	err = db.QueryRow("select username from " + types.DBname + "." + types.UserTable + " where id=?", userId).Scan(&userName)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Get UserName from table users failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Get UserName from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.UserTable)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	var users []types.RetUser
	user := types.RetUser{ID: userId, UserName: userName,Rooms: rooms}
	users = append(users,user)

	ReturnResult=append(ReturnResult,users...)

	temp.UserId=userId
	temp.UserName =userName
	temp.RoomId=roomId
	temp.RoomName = roomName
	temp.ServiceName = serviceName
	temp.ServiceId = serviceId
	temp.TaskNameByUser = service.TaskNameByUser

	// 4. 获取Continers 信息
	containers, err :=  mysql.SelectContainers(db, types.DBname, types.ContainerTable, serviceId)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Select container data from table containers failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select container data from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.ContainerTable)
		ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	ReturnResult[0].Rooms[0].Services[0].Containers = containers


	// 5. 获取 Files 信息
	for kc, container := range ReturnResult[0].Rooms[0].Services[0].Containers{
		files, err :=  mysql.SelectFiles(db, types.DBname, types.FileTable, container.ID)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Select file data from table files failed", true, r, Tracer)
			glog.Error(err)
			glog.Errorf("Select file data from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.FileTable)
			ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		if len(files) != 0 {
			ReturnResult[0].Rooms[0].Services[0].Containers[kc].Files = files
		}else{
			ReturnResult[0].Rooms[0].Services[0].Containers[kc].Files = []types.RetFile{}
		}

		// 6. 获取 Tasks 信息
		for kf, file := range ReturnResult[0].Rooms[0].Services[0].Containers[kc].Files {
			tasks, err := mysql.SelectFileTasks(db, types.DBname, types.TaskTable, file.Id)
			if err != nil {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to Select file data from table tasks failed", true, r, Tracer)
				glog.Error(err)
				glog.Errorf("Select file data from table %s failed in ListServiceTasks(). TracerID: " + TracerId, types.TaskTable)
				ret := types.HttpRespT{HttpCode: http.StatusInternalServerError, Message:"内部错误,查询数据失败", Result: []types.ServiceTask{}}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}

			// 重新将返回结果中时间由秒换算成分钟
			for kt := range tasks {
				tasks[kt].TimeWindow = tasks[kt].TimeWindow / 60
			}

			if len(tasks) != 0 {
				ReturnResult[0].Rooms[0].Services[0].Containers[kc].Files[kf].Tasks = tasks
			} else {
				ReturnResult[0].Rooms[0].Services[0].Containers[kc].Files[kf].Tasks = []types.Task{}
			}
			//count = count + len(tasks)
		}

	}

	temp.Containers = ReturnResult[0].Rooms[0].Services[0].Containers

	RetResult=append(RetResult,temp)
	if len(RetResult) ==0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks due to NO related tasks", false, r, Tracer)
		ret := types.HttpRespT{HttpCode: http.StatusOK, Message:"无相关报警任务", Result: []types.ServiceTask{}}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			panic(err)
		}
		return
	}

	glog.Info("Successful processing the ListServiceTasks request to Logalarm Server")
	ret := types.HttpRespT{HttpCode: http.StatusOK, Message:"获取所有报警任务成功", Result: RetResult}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		panic(err)
	}

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServiceTasks", "Leave ListServiceTasks", false, r, Tracer)
}


/* 根据任务ID删除报警任务 */
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_DeleteTask", "Enter DeleteTask", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit Delete Logalarm Task By TaskID to Logalarm Server. TracerID: " + TracerId)

	var deleteTask types.DeleteTask
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in DeleteTask(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in DeleteTask(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &deleteTask); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in DeleteTask(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	// 参数有效性判断
	if deleteTask.TaskID <=0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Not all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary valid fileds are provide,  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	taskId := deleteTask.TaskID

	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in DeleteTask(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	rows, _:=db.Query("select taskname from " + types.DBname  + "." + types.TaskTable + " where id=?", taskId)
	defer  rows.Close()
	if !rows.Next(){
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to No related tasks found", true, r, Tracer)
		glog.Errorf("No tasks found in [taskid=%d] in DeleteTask(). TracerID: " + TracerId + "\n",taskId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"指定ID任务不存在"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}else{
		var taskName string
		if err := rows.Scan(&taskName); err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to get taskname failed", true, r, Tracer)
			glog.Error(err)
			glog.Error("Failed to get taskname in DeleteTask(). TracerID: " + TracerId)
			ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,获取任务名失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		StringCMD := "/opt/storm/bin/storm kill " + taskName
		if !utils.TopologySubmit(StringCMD, taskName) {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Submit kill Topology task failed", true, r, Tracer)
			glog.Errorf("Submit kill Topology task %s failed with CMD[%s] in DeleteTask(). TracerID: " + TracerId, taskName, StringCMD)
			ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,删除任务失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
	}

	//删除相关任务数据
	tx, _:= db.Begin()
	defer tx.Commit()
	_ ,err =tx.Exec("delete from " +  types.DBname  + "." + types.TaskTable + " where id=?", taskId)
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Leave DeleteTask due to Delete task data in table tasks failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Delete tasks data in table failed in DeleteTask(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,删除任务表数据失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Info("Successful processing the KillTopology request to the Logalarm Server")
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"删除指定ID的任务成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		glog.Error(err)
		return
	}

	// 增加用户操作日志
	//userLog := "删除指定ID报警任务" + taskId
	//utils.RecordUserOperations(taskNameByUser.UserId, userLog)
	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_DeleteTask", "Delete DeleteTask", false, r, Tracer)
}


/* 获取当前用户已创建报警服务的服务列表 */
func ListServices(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_ListServices", "Enter ListServices", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit ListServices request to Logalarm Server. TracerID: " + TracerId)

	var req types.ListServiceRequest
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in ListServices(). TracerID: " + TracerId)
		ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败", Result: []types.ServicesInfo{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in ListServices(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败", Result: []types.ServicesInfo{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in ListServices(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpResS{HttpCode:  http.StatusInternalServerError, Message:"解析请求体失败", Result: []types.ServicesInfo{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	if req.UserId == 0  {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Not all necessary http request parameters are provide", true, r, Tracer)
		glog.Errorf("Not all necessary fileds are provide in ListServices(),  need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段)", Result: []types.ServicesInfo{}}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}


	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in ListServices(). TracerID: " + TracerId)
		ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败", Result: []types.ServicesInfo{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	rows, err := db.Query("select serviceid from " + types.DBname + "." + types.TaskTable + " where userid=" + strconv.Itoa(req.UserId))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Select data from table tasks failed", true, r, Tracer)
		glog.Error(err)
		glog.Errorf("Select data from %s failed in ListServices(). TracerID: " + TracerId, types.TaskTable)
		ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"获取服务ID列表失败", Result: []types.ServicesInfo{}}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	defer rows.Close()

	var serviceIds []int
	for rows.Next() {
		var ServiceId  int

		err := rows.Scan(&ServiceId)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to Try to scan data from table tasks failed", true, r, Tracer)
			glog.Error(err)
			glog.Warningf("Try to scan data from tasks failed in ListServices(). Error info: %s. TracerID: " + TracerId, err)
			ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"读取服务ID列表失败", Result: []types.ServicesInfo{}}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		serviceIds=append(serviceIds, ServiceId)
	}

	var serviceInfo []types.ServicesInfo
	services:=ServiceId_duplicate(serviceIds)
	servicename := ""
	for _, id := range services {
		err = db.QueryRow("select servicename  from " + types.DBname + "." + types.ServiceTable + " where id=?", id).Scan(&servicename)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to select service from table services failed", true, r, Tracer)
			glog.Error(err)
			ret := types.HttpResS{HttpCode: http.StatusInternalServerError, Message:"获取服务名称列表失败", Result: []types.ServicesInfo{}}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		serviceInfo=append(serviceInfo,types.ServicesInfo{ServiceId: id, ServiceName: servicename})
	}

	if len(serviceInfo) == 0{
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices due to No related services", false, r, Tracer)
		ret := types.HttpResS{HttpCode: http.StatusOK, Message:"无相关服务数据", Result: []types.ServicesInfo{}}
		HttpResponse(w, http.StatusOK)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	glog.Infof("Successful processing the ListService [userId= %d] request to Logalarm Server. TracerID: " + TracerId, req.UserId)
	ret := types.HttpResS{HttpCode: http.StatusOK, Message:"获取服务列表成功", Result: serviceInfo}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		return
	}

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_ListServices", "Leave ListServices", false, r, Tracer)
}


/* 停止一组底层（storm 上正在运行的）日志报警任务 */
func StopTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_StopTasks", "Enter StopTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit Stop tasks to Logalarm Server. TracerID: " + TracerId)

	var pendingTasks types.StopStartTask
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in StopTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in StopTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &pendingTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in StopTasks(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 参数有效性判断
	if len(pendingTasks.TaskNames) == 0 || pendingTasks.UserId <=0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to No tasknames are provide", true, r, Tracer)
		glog.Errorf("No tasknames are provide in StopTasks(), need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	for _,taskName := range pendingTasks.TaskNames {
		// 名称合法性
		if strings.Contains(taskName.TaskName, "."){
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to taskname contains invalid character", true, r, Tracer)
			glog.Errorf("Taskname %s contains invalid character in StopTasks(), bad request Body. TracerID: " + TracerId, taskName.TaskName)
			HttpResponse(w, http.StatusInternalServerError)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(任务名称包含特殊字符\".\")"}
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
	}

	// 查看任务是否存在
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in StopTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	for _,taskName := range pendingTasks.TaskNames {

		tasks, err := mysql.SelectTasksByTaskName(db, types.DBname, types.TaskTable, taskName.TaskName)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Get task from table tasks failed", true, r, Tracer)
			glog.Error(err)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查找任务失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		/*
		rows, err := db.Query("select id,taskstate from " + types.DBname + "." + types.TaskTable + " where taskname='" + taskName.TaskName + "'")
		if err != nil {
			glog.Error(err)
			glog.Errorf("Select data from %s failed in StopTasks()", types.TaskTable)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"获取服务ID列表失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		defer rows.Close()

		var taskIds []int
		for rows.Next() {
			var taskId  int

			err := rows.Scan(&taskId)
			if err != nil {
				glog.Error(err)
				glog.Warningf("Try to scan data from tasks failed in StopTasks(). Error info: %s. ", err)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"读取服务ID列表失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
			taskIds=append(taskIds, taskId)
		}
		if len(taskIds) == 0 {
			glog.Infof("No matched taskname found for %s in database in StartTasks().", taskName.TaskName)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError,Message:"无匹配的taskname"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		*/

		if len(tasks) == 0 {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to No matched taskname found", true, r, Tracer)
			glog.Infof("No matched taskname found for %s in database in StopTasks(). TracerID: " + TracerId, taskName.TaskName)
			ret := types.HttpRe{HttpCode: http.StatusOK,Message:"无匹配的taskname"}
			HttpResponse(w, http.StatusOK)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		for _, task := range tasks{
			if task.TaskState != "active" {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Matched task is not running at active state", true, r, Tracer)
				glog.Infof("Matched taskname %s in database is not running at active state in StopTasks(). TracerID: " + TracerId, task.TaskName)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError,Message:"任务状态已经是inactive"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
		}
	}

	// 停止任务
	tx,_ := db.Begin()
	defer tx.Commit()
	for _,taskName := range pendingTasks.TaskNames{
		StringCMD := "/opt/storm/bin/storm deactivate " + taskName.TaskName
		if !utils.TopologySubmit(StringCMD, taskName.TaskName) {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Submit deactive task failed", true, r, Tracer)
			glog.Errorf("Submit deactive task %s failed with CMD[%s] in StopTasks(). TracerID: " + TracerId, taskName.TaskName, StringCMD)
			ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,停止任务失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		if ! mysql.Update_TaskStatus(tx, types.DBname, types.TaskTable, taskName.TaskName , "inactive") {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks due to Update taskstate failed", true, r, Tracer)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"停止任务失败(更新报警任务状态失败)"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				glog.Error(err)
				return
			}
			return
		}

		// 增加用户操作日志
		userLog := "停止报警任务" + taskName.TaskName
		utils.RecordUserOperations(pendingTasks.UserId, userLog)
	}
	tx.Commit()

	glog.Info("Successful processing the StopTasks request to the Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"停止任务成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		glog.Error(err)
		return
	}

	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StopTasks", "Leave StopTasks", false, r, Tracer)
	return
}


/* 重新启动一组底层（storm 上 deactive 的）日志报警任务 */
func StartTasks(w http.ResponseWriter, r *http.Request) {
	var Tracer *tracing.DHTracer
	// traceing log
	Tracer,_ = logtracer.TracerAndSpan_Entry("Logalarmserver_StartTasks", "Enter StartTasks", r, Tracer)
	TracerId := logtracer.GetLogTracerID(Tracer.Tracer, Tracer.ActiveSpan)

	glog.Info("Submit Start tasks to Logalarm Server. TracerID: " + TracerId)

	var pendingTasks types.StopStartTask
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Read http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Read request body failed in StartTasks()")
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := r.Body.Close(); err != nil {
		glog.Error(err)
		glog.Error("Close request body reading failed in StartTask(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,关闭读取请求失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}
	if err := json.Unmarshal(body, &pendingTasks); err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Unmarshal the http requestBody failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Unmarshal the request body failed in StartTask(). TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	// 参数有效性判断
	if len(pendingTasks.TaskNames) == 0 || pendingTasks.UserId <=0 {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to No tasknames provided", true, r, Tracer)
		glog.Errorf("No tasknames are provide in StartTasks(), need check request Body. TracerID: " + TracerId)
		HttpResponse(w, http.StatusInternalServerError)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(未提供必要的字段或者字段值非法)"}
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	for _,taskName := range pendingTasks.TaskNames {
		// 名称合法性
		if strings.Contains(taskName.TaskName, "."){
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to taskname contains invalid character(.)", true, r, Tracer)
			glog.Errorf("Taskname %s contains invalid character in StartTasks(), bad request Body. TracerID: " + TracerId, taskName.TaskName)
			HttpResponse(w, http.StatusInternalServerError)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"请求格式错误(任务名称包含特殊字符\".\")"}
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
	}

	// 查看任务是否存在
	db, err := mysql.ConnectMYSQL()
	defer  db.Close()
	if err != nil {
		Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Connect to Mysql failed", true, r, Tracer)
		glog.Error(err)
		glog.Error("Connect to Mysql failed in StartTasks(). TracerID: " + TracerId)
		ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,连接Mysql数据库失败"}
		HttpResponse(w, http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(ret); err != nil {
			return
		}
		return
	}

	for _,taskName := range pendingTasks.TaskNames {
		tasks, err := mysql.SelectTasksByTaskName(db, types.DBname, types.TaskTable, taskName.TaskName)
		if err != nil {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Get task from table tasks failed", true, r, Tracer)
			glog.Error(err)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"内部错误,查找任务失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		/*
		rows, err := db.Query("select id from " + types.DBname + "." + types.TaskTable + " where taskname='" + taskName.TaskName + "'")
		if err != nil {
			glog.Error(err)
			glog.Errorf("Select data from %s failed in StartTasks()", types.TaskTable)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"获取服务ID列表失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}
		defer rows.Close()

		var taskIds []int
		for rows.Next() {
			var taskId  int

			err := rows.Scan(&taskId)
			if err != nil {
				glog.Error(err)
				glog.Warningf("Try to scan data from tasks failed in StartTasks(). Error info: %s. ", err)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"读取服务ID列表失败"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
			taskIds=append(taskIds, taskId)
		}

		if len(taskIds) == 0 {
			glog.Infof("No matched taskname found for %s in database in StartTasks().", taskName.TaskName)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError,Message:"无匹配的taskname"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}*/

		if len(tasks) == 0 {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to No matched taskname found", true, r, Tracer)
			glog.Infof("No matched taskname found for %s in database in StartTasks(). TracerID: " + TracerId, taskName.TaskName)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError,Message:"无匹配的taskname"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		for _, task := range tasks{
			if task.TaskState != "inactive" {
				Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Matched task is not running at inactive state", true, r, Tracer)
				glog.Infof("Matched taskname %s in database is not running at inactive state in StartTasks(). TracerID: " + TracerId, task.TaskName)
				ret := types.HttpRe{HttpCode: http.StatusInternalServerError,Message:"任务状态已经是active"}
				HttpResponse(w, http.StatusInternalServerError)
				if err := json.NewEncoder(w).Encode(ret); err != nil {
					return
				}
				return
			}
		}
	}

	tx,_ := db.Begin()
	defer tx.Commit()
	// 启动任务
	for _,taskName := range pendingTasks.TaskNames{
		StringCMD := "/opt/storm/bin/storm activate " + taskName.TaskName
		if !utils.TopologySubmit(StringCMD, taskName.TaskName) {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Submit activate task failed", true, r, Tracer)
			glog.Errorf("Submit activate task %s failed with CMD[%s] in StartTasks(). TracerID: " + TracerId, taskName.TaskName, StringCMD)
			ret := types.HttpRe{HttpCode:  http.StatusInternalServerError, Message:"内部错误,启动任务失败"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				return
			}
			return
		}

		if ! mysql.Update_TaskStatus(tx, types.DBname, types.TaskTable, taskName.TaskName , "active") {
			Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks due to Update taskstate failed", true, r, Tracer)
			ret := types.HttpRe{HttpCode: http.StatusInternalServerError, Message:"停止任务失败(更新报警任务状态失败)"}
			HttpResponse(w, http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(ret); err != nil {
				glog.Error(err)
				return
			}
			return
		}

		// 增加用户操作日志
		userLog := "重新启动报警任务" + taskName.TaskName
		utils.RecordUserOperations(pendingTasks.UserId, userLog)
	}
	tx.Commit()

	glog.Info("Successful processing the DeactiveTopology request to the Logalarm Server. TracerID: " + TracerId)
	ret := types.HttpRe{HttpCode: http.StatusOK, Message:"启动任务成功"}
	HttpResponse(w, http.StatusOK)
	if err := json.NewEncoder(w).Encode(ret); err != nil {
		glog.Error(err)
		return
	}
	Tracer = logtracer.TracerAndSpan_Leave("Logalarmserver_StartTasks", "Leave StartTasks", false, r, Tracer)
	return
}

// 从任务表查找到的serviceid 有重复，需要删除
func ServiceId_duplicate(list []int) []int {
	var x []int = []int{}
	for _, i := range list {
		if len(x) == 0 {
			x = append(x, i)
		} else {
			for k, v := range x {
				if i == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, i)
				}
			}
		}
	}
	return x
}


func HttpResponse(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
}

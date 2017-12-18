package handler

import (
	"fmt"
	"net/http"
//	"github.com/gorilla/mux"
	"github.com/wangzhuzhen/logalarmserver/types"
	"encoding/json"
	"io/ioutil"
	"io"
	"github.com/wangzhuzhen/logalarmserver/mysql"
	"github.com/wangzhuzhen/logalarmserver/submittopoloy"
)

func StatusCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to Logalarm server!")
}


func ListRules(w http.ResponseWriter, r *http.Request) {
	/*
	rules := types.Rules{
		types.Rule{RuleOwner:"wangzhuzhen",RuleName: "Rule001", KeyWord: "ERROR", KeywordIndex: 8, TimeWindow: 30, ThresholdNum: 5, EmailList: "wang_zhuzhen@dahuatech.com"},
		types.Rule{RuleOwner:"wangzhuzhen",RuleName: "Rule002", KeyWord: "FATAL", KeywordIndex: 7, TimeWindow: 10, ThresholdNum: 2, EmailList: "wang_zhuzhen@dahuatech.com"},
	}

	fmt.Printf("Input value rules: %v \n", rules)
	*/

	db, err := mysql.ConnectMYSQL()
	if err != nil {
		panic(err)
	}

	rules, err := mysql.SelectData_Rules(db, types.DBname)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

func ListUserRules(w http.ResponseWriter, r *http.Request) {

	var user types.RuleUser
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &user); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity


		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	db, err := mysql.ConnectMYSQL()
	if err != nil {
		panic(err)
	}

	rules, err := mysql.Select_UserRules(db, types.DBname, user)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(rules); err != nil {
		panic(err)
	}
}

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
	if err != nil {
		panic(err)
	}

	mysql.CreteDatabase(db, types.DBname)

	mysql.CreteTable_Rules(db, types.DBname)
	mysql.InsertData_Rules(db, types.DBname,rule)
/*
	mysql.CreteTable_Topology(db, "wang")
	mysql.InsertData_Topology(db, "wang")

	mysql.UpdateData_Rules(db, "wang")
	mysql.UpdateData_Topology(db, "wang")

	mysql.SelectData_Rules(db, "wang")
	mysql.SelectData_Topology(db, "wang")

	//mysql.DeleteRules(db, "wang")
	//mysql.DeleteTopology(db, "wang")

	//StringCMD := "echo hello world"
*/
	StringCMD := "echo test submit topology"
	submittopoloy.SubmitTopolgy(StringCMD)
}

func UpdateRule(w http.ResponseWriter, r *http.Request) {

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
	/*
	if err := json.NewEncoder(w).Encode(rule); err != nil {
		panic(err)
	}
	*/
	db, err := mysql.ConnectMYSQL()
	if err != nil {
		panic(err)
	}

	mysql.UpdateData_Rule(db, types.DBname,rule)
}


func DeleteRule(w http.ResponseWriter, r *http.Request) {

	var drule types.DeletedRule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &drule); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	db, err := mysql.ConnectMYSQL()
	if err != nil {
		panic(err)
	}

	mysql.DeleteRule(db, types.DBname, drule)
}


package handler

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/wangzhuzhen/logalarmserver/types"
	"encoding/json"
	"io/ioutil"
	"io"
	"github.com/wangzhuzhen/logalarmserver/mysql"
	"github.com/wangzhuzhen/logalarmserver/submittopoloy"
)

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}


func TodoShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["todoId"]
	fmt.Fprintln(w, "Todo show:", todoId)
}

func ListRules(w http.ResponseWriter, r *http.Request) {
	todos := types.Rules{
		types.Rule{RuleName: "Rule ABC"},
		types.Rule{RuleName: "Rule EFG"},
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		panic(err)
	}
}

func CreateRule(w http.ResponseWriter, r *http.Request) {
	var todo types.Rule
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &todo); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

//	t := RepoCreateTodo(todo)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(todo); err != nil {
		panic(err)
	}
	db, err := mysql.ConnectMYSQL()
	if err != nil {
		panic(err)
	}

	mysql.CreteDatabase(db, "wang")

	mysql.CreteTable_Rules(db, "wang")
	mysql.InsertData_Rules(db, "wang")

	mysql.CreteTable_Topology(db, "wang")
	mysql.InsertData_Topology(db, "wang")

	mysql.UpdateData_Rules(db, "wang")
	mysql.UpdateData_Topology(db, "wang")

	mysql.SelectData_Rules(db, "wang")
	mysql.SelectData_Topology(db, "wang")

	mysql.DeleteData_Rules(db, "wang")
	mysql.DeleteData_Topology(db, "wang")

	//StringCMD := "echo hello world"
	StringCMD := "cat /etc/hosts /etc/resolv.conf"
	submittopoloy.SubmitTopolgy(StringCMD)
}


package types

import (
	"net/http"
)

var DBname string="wang"

type Rule struct {
	RuleOwner      string        `json:"ruleowner"`
	RuleName       string        `json:"rulename"`
	KeyWord        string        `json:"keyword"`
	KeywordIndex   int           `json:"keywordindex"`
	TimeWindow     int           `json:"timewindow"`
	ThresholdNum   int           `json:"thresholdnum"`
	EmailList     string	     `json:"emaillist"`
}

type Rules []Rule

type DeletedRule struct {
	RuleOwner string        `json:"ruleowner"`
	RuleName  string        `json:"rulename"`
}

type RuleUser struct {
	Username string        `json:"ruleowner"`
}

type Topology struct {
	ToplogyName      string      `json:"topologyname"`
	AppName          string      `json:"appname"`
	Submitted         int         `json:"submit"`
	Rule
}


type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route
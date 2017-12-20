package types

import (
	"net/http"
)

/* 报警数据库名称 */
var DBname string="wang"

/* 日志报警规则表 */
type Rule struct {
	RuleOwner      string        `json:"ruleowner"`
	RuleName       string        `json:"rulename"`
	KeyWord        string        `json:"keyword"`
	KeywordIndex   int           `json:"keywordindex"`
	TimeWindow     int           `json:"timewindow"`
	ThresholdNum   int           `json:"thresholdnum"`
	EmailList     string	     `json:"emaillist"`
}

/* 日志报警规则的规则所有者和规则名称 */
type RuleUser struct {
	RuleOwner string        `json:"ruleowner"`
	RuleName  string        `json:"rulename"`
}

/* 日志报警任务表 */
type Topology struct {
	TopologyOwner     string      `json:"topologyowner"`
	TopologyName      string      `json:"topologyname"`
	AppName           string      `json:"appname"`
	Submitted         int         `json:"submitted"`
	RuleOwner         string      `json:"ruleowner"`
	RuleName          string      `json:"rulename"`
	KeyWord           string      `json:"keyword"`
	KeywordIndex      int         `json:"keywordindex"`
	TimeWindow        int         `json:"timewindow"`
	ThresholdNum      int         `json:"thresholdnum"`
	EmailList         string      `json:"emaillist"`
}

/* 日志报警任务的所有者和任务名称 */
type TopologyUser struct {
	TopologyOwner     string      `json:"topologyowner"`
	TopologyName      string      `json:"topologyname"`
}

/* HTTP 路由信息 */
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

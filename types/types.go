package types

import (
	"net/http"
)

/* 报警数据库名称 */
var DBname string="wang"

/* 日志报警规则表 */
type Rule struct {
	RuleOwner      string        `json:"ruleOwner"`
	RuleName       string        `json:"ruleName"`
	KeyWord        string        `json:"keyword"`
	KeywordIndex   int           `json:"keywordIndex"`
}

/* 日志报警规则的规则所有者和规则名称 */
type RuleUser struct {
	RuleOwner string        `json:"ruleOwner"`
	RuleName  string        `json:"ruleName"`
}

/* 日志报警任务表 */
type Topology struct {
	TopologyOwner     string      `json:"topologyOwner"`
	TopologyName      string      `json:"topologyName"`
	AppName           string      `json:"appName"`
	KeyWord           string      `json:"keyword"`
	KeywordIndex      int         `json:"keywordIndex"`
	TimeWindow        int         `json:"timeWindow"`
	ThresholdNum      int         `json:"thresholdNum"`
	EmailList         string      `json:"emailList"`
}

/* 日志报警任务的所有者和任务名称 */
type TopologyUser struct {
	TopologyOwner     string      `json:"topologyOwner"`
	TopologyName      string      `json:"topologyName"`
}

/* HTTP 路由信息 */
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

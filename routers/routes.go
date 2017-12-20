package routers

import (
	"github.com/wangzhuzhen/logalarmserver/handler"
	"github.com/wangzhuzhen/logalarmserver/types"
)

type Routes []types.Route

//var routes = types.Routes{
var routes = Routes{
	types.Route{
		"StatusCheck",    /* 服务状态检查 */
		"GET",
		"/",
		handler.StatusCheck,
	},
	types.Route{
		"ListRules",    /* 查看所有报警规则 */
		"GET",
		"/rules",
		handler.ListRules,
	},
	types.Route{
		"ListUserRules",   /* 查看指定用户的所有报警规则 */
		"GET",
		"/rules/{ruleowner}",
		handler.ListUserRules,
	},
	types.Route{
		"CreateRule",    /* 创建新的报警规则 */
		"POST",
		"/rules",
		handler.CreateRule,
	},
	types.Route{
		"UpdateRule",      /* 更新指定用户指定规则名称的报警规则 */
		"POST",
		"/rules/{ruleowner}/{rulename}",
		handler.UpdateRule,
	},
	types.Route{
		"DeleteRule",    /* 删除指定用户指定规则名称的报警规则 */
		"DEL",
		"/rules/{ruleowner}/{rulename}",
		handler.DeleteRule,
	},
	types.Route{
		"ListTopologys",     /* 查看所有报警任务 */
		"GET",
		"/topologys",
		handler.ListTopologys,
	},
	types.Route{
		"ListUserTopologys",    /* 查看指定用户的所有报警任务 */
		"GET",
		"/topologys/{topologyowner}",
		handler.ListUserTopologys,
	},
	types.Route{
		"CreateTopology",     /* 创建新的报警任务 */
		"POST",
		"/topologys",
		handler.CreateTopology,
	},
	types.Route{
		"UpdateTopology",      /* 更新指定用户指定任务名称的报警任务 */
		"POST",
		"/topologys/{topologyowner}/{topologyname}",
		handler.UpdateTopology,
	},
	types.Route{
		"DeleteTopology",   /* 删除指定用户指定报警任务名称的报警任务 */
		"DEL",
		"/topologys/{topologyowner}/{topologyname}",
		handler.DeleteTopology,
	},
	types.Route{
		"SubmitTopology",   /* 提交指定用户指定名称的报警任务 */
		"POST",
		"/topologys/{topologyowner}/{topologyname}/submit",
		handler.SubmitTopology,
	},
}


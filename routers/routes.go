package routers

import (
	"github.com/wangzhuzhen/logalarmserver/handler"
	"github.com/wangzhuzhen/logalarmserver/types"
)

var routes = types.Routes{
	types.Route{
		"StatusCheck",
		"GET",
		"/",
		handler.StatusCheck,
	},
	types.Route{
		"ListRules",
		"GET",
		"/rules",
		handler.ListRules,
	},
	types.Route{
		"ListUserRules",
		"GET",
		"/rules/{ruleowner}",
		handler.ListUserRules,
	},
	types.Route{
		"UpdateRules",
		"POST",
		"/rules/{ruleowner}/{rulename}",
		handler.UpdateRule,
	},
	types.Route{
		"CreateRule",
		"POST",
		"/rules",
		handler.CreateRule,
	},
	types.Route{
		"DeleteRule",
		"DEL",
		"/rules/{ruleowner}/{rulename}",
		handler.DeleteRule,
	},
}


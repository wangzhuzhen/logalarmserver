package routers

import (
	"github.com/wangzhuzhen/logalarmserver/handler"
	"github.com/wangzhuzhen/logalarmserver/types"
)

var routes = types.Routes{
	types.Route{
		"Index",
		"GET",
		"/",
		handler.Index,
	},
	types.Route{
		"ListRules",
		"GET",
		"/rules",
		handler.ListRules,
	},
	types.Route{
		"TodoShow",
		"GET",
		"/todos/{todoId}",
		handler.TodoShow,
	},
	types.Route{
		"CreateRule",
		"POST",
		"/rules",
		handler.CreateRule,
	},
}


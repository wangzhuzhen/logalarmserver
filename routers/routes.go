package routers

import (
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/handler"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/types"
)

type Routes []types.Route

var routes = Routes{
	types.Route{
		"StatusCheck",    /* 服务状态检查 */
		"GET",
		"/",
		handler.StatusCheck,
	},
	types.Route{
		"CreateRule",    /* 创建新的报警规则 */
		"POST",
		"/createrule",
		handler.CreateRule,
	},
	types.Route{
		"ListRules",    /* 查看所有报警规则 */
		"POST",
		"/listrules",
		handler.ListRules,
	},
	types.Route{
		"ListSingleRules",   /* 查看指定用户的所有报警规则 */
		"POST",
		"/listrule",
		handler.ListSingleRule,
	},
	types.Route{
		"UpdateRule",      /* 更新指定用户指定规则名称的报警规则 */
		"POST",
		"/updaterule",
		handler.UpdateRule,
	},
	types.Route{
		"DeleteRule",    /* 删除指定用户指定规则名称的报警规则 */
		"POST",
		"/deleterule",
		handler.DeleteRule,
	},
	types.Route{
		"SubmitTasks",   /* 以服务为单位提交提交报警任务 */
		"POST",
		"/submittasks",
		handler.SubmitTasks,
	},
	types.Route{
		"ListTasks",     /* 查询报警任务，如果传入参数包含用户ID，则是查询用户的所有报警任务，否则是平台的所有报警任务 */
		"POST",
		"/listtasks",
		handler.ListTasks,
	},
	types.Route{
		"ListServiceTasks",    /* 查看指定服务的所有报警任务 */
		"POST",
		"/listservicetasks",
		handler.ListServiceTasks,
	},
	types.Route{
		"UpdateTasks",     /* 以服务为单位更新的报警任务 */
		"POST",
		"/updateservicetasks",
		handler.UpdateServiceTasks,
	},
	types.Route{
		"StopTasks",     /* 暂时关闭/停止相关报警任务列表 */
		"POST",
		"/stoptasks",
		handler.StopTasks,
	},
	types.Route{
		"StartTasks",     /* 重新启动之前关闭/停止的报警任务列表 */
		"POST",
		"/starttasks",
		handler.StartTasks,
	},
//	types.Route{
//		"UpdateTask",      /* 更新指定ID的报警任务 */
//		"POST",
//		"/updatetask/{taskid}",
//		handler.UpdateTask,
//	},
	types.Route{
		"DeleteTask",   /* 删除指定ID的报警任务 */
		"POST",
		"/deletetask",
		handler.DeleteTask,
	},
	types.Route{
		"DeleteTasks",   /* 删除指定服务ID的所有报警任务 */
		"POST",
		"/deleteservicetasks",
		handler.DeleteServiceTasks,
	},
	types.Route{
		"ListServices",   /* 根据用户名获取用户已经添加的报警任务的服务列表 */
		"POST",
		"/listservices",
		handler.ListServices,
	},
}


package types

import (
	"net/http"
)

const (
	DBname          string="logalarm"         // 数据库名称
	RuleTable        string="rules"            // 报警规则表名称
	TaskTable        string="tasks"            // 报警任务表名称
//	TaskInfoTable    string="taskinfo"         // 报警服务用户表名称
	FileTable        string="files"            // 报警文件表名称
	ContainerTable   string="containers"       // 报警容器表名称
	ServiceTable     string="services"         // 报警服务表名称
	RoomTable        string="rooms"            // 报警任务的room表名称
	UserTable        string="users"            // 报警服务用户表名称
	TaskNameTable    string="tasknames"        // 用户以服务为单位提交报警任务时指定的整个服务的报警任务的名称


	// 表创建CMD
	// create table if not exists ruletest(id int auto_increment, rulename varchar(255), userid int, username varchar(255), keyword varchar(255), keywordindex int, createtimestamp bigint, updatetimestamp bigint, primary key(id), unique(id));
	RuleTableCreateCMD string="create table if not exists  "+ DBname + "." + RuleTable +"(id int auto_increment, rulename varchar(255), userid int, username varchar(255), keyword varchar(255), keywordindex int, createtimestamp bigint, updatetimestamp bigint, primary key(id), unique(id))"
	// create table if not exists usertest(id int, username varchar(255), createtimestamp bigint, primary key(id), unique(id));
	UserTableCreateCMD string="create table if not exists  "+ DBname + "." + UserTable +"(id int, username varchar(255),createtimestamp bigint, primary key(id), unique(id))"
	// create table if not exists roomtest(id int, roomname varchar(255), uid int, createtimestamp bigint, primary key(id), unique(id));
	RoomTableCreateCMD string="create table if not exists  "+ DBname + "." + RoomTable +"(id int, roomname varchar(255), uid int, createtimestamp bigint, primary key(id), unique(id))"
	// create table if not exists servicetest(id int, servicename varchar(255), rid int, taskname varchar(255), createtimestamp bigint, primary key(id), unique(id));
	ServiceTableCreateCMD string="create table if not exists  "+ DBname + "." + ServiceTable +"(id int, servicename varchar(255), rid int, taskname varchar(255), createtimestamp bigint, primary key(id), unique(id))"
	// create table if not exists containertest(id int, containername varchar(255), sid int, createtimestamp bigint, primary key(id), unique(id));
	ContainerTableCreateCMD string="create table if not exists  "+ DBname + "." + ContainerTable +"(id int, containername varchar(255), sid int, createtimestamp bigint, primary key(id), unique(id))"
	// create table if not exists filetest(id int auto_increment, filename varchar(255), filepath varchar(2047), cid int, sid int,createtimestamp bigint, primary key(id),unique(id));
	FileTableCreateCMD string="create table if not exists  "+ DBname + "." + FileTable +"(id int auto_increment, filename varchar(255), filepath varchar(2047), cid int,sid int, createtimestamp bigint, primary key(id),unique(id))"
	// create table if not exists tasktest(id int auto_increment, taskname varchar(2047), ruleid int, rulename varchar(255), keyword varchar(255), keywordindex int, timewindow int, number int, groupid int, groupname varchar(255), filepath varchar(2047),filename varchar(255), fileid int, serviceid int, containerid int, roomid int, userid int,createtimestamp bigint, updatetimestamp bigint, taskstate varchar(255), primary key(id), unique(taskname, id));
	TaskTableCreateCMD string="create table if not exists  "+ DBname + "."  + TaskTable+ "(id int auto_increment, taskname varchar(2047), ruleid int, rulename varchar(255), keyword varchar(255), keywordindex int, timewindow int, number int, groupid int, groupname varchar(255), filepath varchar(2047),filename varchar(255), fileid int, serviceid int, containerid int, roomid int, userid int, createtimestamp bigint, updatetimestamp bigint, taskstate varchar(255), primary key(id), unique(taskname, id))"
	// create table if not exists tasknametest(id int auto_increment, userid int, serviceid int, taskname varchar(255), primary key(id));
	TaskNameTableCreateCMD string="create table if not exists  "+ DBname + "." + TaskNameTable +"(id int auto_increment, userid int, serviceid int, taskname varchar(255), primary key(id))"

	// 表数据插入CMD
	// insert into ruletest(rulename,userid,username,keyword,keywordindex,createtimestamp,updatetimestamp) values("rule001",1,"wang001","ERROR",8,11111000111,11111000111);
	RuleTableInsertElements string="insert into " + DBname + "." + RuleTable + "(rulename,userid,username,keyword,keywordindex,createtimestamp,updatetimestamp) values(?,?,?,?,?,?,?)"
	// insert into usertest(id,username,createtimestamp) values(2,"wangzhuzhen",11112222000);
	UserTableInsertElements string="insert into " + DBname + "." + UserTable + "(id,username,createtimestamp) values(?,?,?)"
	// insert into roomtest(id,roomname,uid,createtimestamp) values(2,"wangzhuzhen",3,11112222000);
	RoomTableInsertElements string="insert into " + DBname + "." + RoomTable + "(id,roomname,uid,createtimestamp) values(?,?,?,?)"
	// insert into servicetest(id,servicename,rid,taskname,createtimestamp) values(1,"wong",3,"task001",1122334455667);
	ServiceTableInsertElements string="insert into " + DBname + "." + ServiceTable + "(id,servicename,rid,taskname,createtimestamp) values(?,?,?,?,?)"
	// insert into containertest(id,containername,sid,createtimestamp) values(5,"wongc",2,1234567890123)
	ContainerTableInsertElements string="insert into " + DBname + "." + ContainerTable + "(id,containername,sid,createtimestamp) values(?,?,?,?)"
	// insert into filetest(filename,filepath,cid,sid,createtimestamp) values("test.abc.log","/va/log/dhc/applog/",3,2,1567982340125)
	FileTableInsertElements string="insert into " + DBname + "." + FileTable + "(filename,filepath,cid,sid,createtimestamp) values(?,?,?,?,?)"
	// insert into tasktest(taskname,ruleid,rulename,keyword,keywordindex,timewindow,number,groupid,groupname,filepath,filename,fileid,serviceid,containerid,roomid,userid,createtimestamp,updatetimestamp,taskstate) values("abc.efg.hij.kl.mno",1,"rule001","ERROR",8,100,10,1,"GROUPWONG","/var/log/app/","kl",10,2,1,3,3,1234567899874,1234567899874,"active");
	TaskTableInsertElements string="insert into " + DBname + "." + TaskTable +"(taskname,ruleid,rulename,keyword,keywordindex,timewindow,number,groupid,groupname,filepath,filename,fileid,serviceid,containerid,roomid,userid,createtimestamp,updatetimestamp,taskstate) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	// insert into tasknametest(userid,serviceid,taskname) values(2,2,"taskname001");
	TaskNameTableInsertElements string="insert into " + DBname + "." + TaskNameTable + "(userid,serviceid,taskname) values(?,?,?)"

	// 表数据存在性查询CMD
	RuleDataExistedCMD string="select count(*) from " +  DBname + "." + RuleTable + " where "
	UserDataExistedCMD string="select count(*) from " +  DBname + "." + UserTable + " where id="
	RoomDataExistedCMD string="select count(*) from " +  DBname + "." + RoomTable + " where id="
	ServiceDataExistedCMD string="select count(*) from " +  DBname + "." + ServiceTable + " where id="
	ContainerDataExistedCMD string="select count(*) from " +  DBname + "." + ContainerTable + " where id="
	FileDataExistedCMD string="select count(*) from " +  DBname + "." + FileTable + " where "
	TaskDataExistedCMD string="select count(*) from " +  DBname + "." + TaskTable + " where taskname='"
	TaskNameDataExistedCMD string="select count(*) from " +  DBname + "." + TaskNameTable + " where userid="


	UserOperationLogDir string="/var/log/dhc/operationlog"
	UserOperationLogFile string="logalarm-user-operation.log"

)


// 1. 请求返回部分相关结构
/* 不含结果的 Http 请求返回信息 */
type HttpRe struct {
	HttpCode  int             `json:"code"`       // 返回码
	Message   string        `json:"message"`    // 返回信息
}

/* Http 请求返回单条信息 */
type HttpR struct{
	HttpCode    int          `json:"code"`       // 返回码
	Message     string       `json:"message"`    // 返回信息
	Result      Rule         `json:"result"`     // HTTP请求返回结果
}

/* Http 请求返回多条信息 */
type HttpRes struct{
	HttpCode    int          `json:"code"`       // 返回码
	Message     string       `json:"message"`    // 返回信息
	Result      []Rule         `json:"result"`     // HTTP请求返回结果
}

/* Http 请求返回多条信息 */
type HttpResS struct{
	HttpCode    int          `json:"code"`       // 返回码
	Message     string       `json:"message"`    // 返回信息
	Result      []ServicesInfo         `json:"result"`     // HTTP请求返回结果
}

/* 报警规则 Http list 请求返回多条信息 */
type HttpRespR struct{
	HttpCode    int          `json:"code"`       // 返回码
	Message     string       `json:"message"`    // 返回信息
	Result      []Rule       `json:"result"`     // HTTP请求返回结果
	TotalPages  int          `json:"totalPages"`  // 所有结果占的页数

}

/* 报警任务 Http list 请求返回多条信息 */
type HttpRespT struct{
	HttpCode    int          `json:"code"`       // 返回码
	Message     string       `json:"message"`    // 返回信息
	Result      []ServiceTask /*RetUser*/   `json:"result"`     // HTTP请求返回结果
	TotalPages  int          `json:"totalPages"`  // 所有结果占的页数

}



// 2. 报警联系人相关信息
/* 报警联系人组的 contact 内容用于发送报警邮件 */
type Contact struct {
	Mobile             string     `json:"mobile"`
	Mail               string     `json:"mail"`
	//Name               string     `json:"name"`
	//UserId             int        `json:"userId"`
	//Id                 int        `json:"id"`
	//CreateTime         string     `json:"CreateTime"`
	//UpdateTime         string     `json:"updateTime"`
	//CreateTimestamp    int64      `json:"createTimestamp"`
	//UpdateTimestamp    int64      `json:"updateTimestamp"`
}

/* 存储报警任务组的邮箱和收集列表信息 */
type PhonesAndEmails struct {
	GroupName          string       `json:"name"`
	GroupId            int          `json:"id"`
	Contacts           []Contact    `json:"contact"`
}


// 3. List 规则表和任务表用到的请求体
/* 报警任务表、规则表查表请求 */
type ListRequest struct{
	CurrentPage     int      `json:"currentPage"`    // 请求查的页编号
	PageSize        int      `json:"pageSize"`       // 每页记录的数据条数
	UserId          int      `json:"userId"`         // 规则的所有者(用户)的ID，为空则表示查询所有用户的所有规则
}



// 4. List 任务表用到的相关结果
/* 以服务为单位描述的任务信息，用于以服务为单位返回任务列表 */
type ServiceTask struct {
	TaskNameByUser string       `json:"taskName"`        // 用户指定的服务为单位提交的任务名称
	UserName    string     `json:"userName,omitempty"`       // 任务所属的服务的user名称
	UserId      int        `json:"userId,omitempty"`       // 任务所属的服务的user的 ID
	RoomName    string     `json:"roomName,omitempty"`       // 任务所属的服务的Room名称
	RoomId      int        `json:"roomId,omitempty"`       // 任务所属的服务的Room的 ID
	ServiceName    string     `json:"serviceName,omitempty"`       // 任务所属的服务的名称
	ServiceId      int        `json:"serviceId,omitempty"`       // 任务所属的服务的 ID
	Containers     []RetContainer   `json:"containers,omitempty"`
}

/* 用于以服务为单位返回任务列表时记录User信息 */
type RetUser struct {
	ID            int           `json:"id"`           // 页面传入的用户ID
	UserName      string        `json:"userName"`     // Room 的名称
	Rooms	      []RetRoom        `json:"rooms"`
}
/* 用于以服务为单位返回任务列表时记录Room信息 */
type RetRoom struct {
	ID    int             `json:"roomId"`
	RoomName  string          `json:"roomName"`
	Services  []RetService    `json:"services"`
}
/* 用于以服务为单位返回任务列表时记录Service信息 */
type RetService struct {
	 ID      int              `json:"serviceId"`
	 ServiceName    string           `json:"serviceName"`
	 TaskNameByUser string       `json:"taskName,omitempty"`        // 用户指定的服务为单位提交的任务名称
	 Containers     []RetContainer   `json:"containers"`
 }
/* 用于以服务为单位返回任务列表时记录Container信息 */
type RetContainer struct {
	ID     int              `json:"containerId,omitempty"`
	ContainerName    string           `json:"containerName,omitempty"`
	Files            []RetFile        `json:"files"`
}
/* 用于以服务为单位返回任务列表时记录File信息 */
type RetFile struct {
	Id            int           `json:"fileId"`           // 报警日志文件在日志文件表中的ID，每插入一条自增
	FileName      string        `json:"fileName"`         // 报警文件名称
	FilePath      string        `json:"filePath"`         // 报警文件的路径名
	ContainerId   int           `json:"containerId,omitempty"`      // 报警文件所属的容器在容器表中的ID
	ServiceId   int           `json:"serviceId,omitempty"`      // 报警文件所属服务的ID
	Tasks         []Task        `json:"tasks"`
}



// 5. 创建各种表用到的结构

/* 日志报警服务维护的报警规则表 */
type Rule struct {
	Id             int           `json:"ruleId,omitempty"`           // 报警规则的ID，每插入一条自增
	RuleName       string        `json:"ruleName,omitempty"`         // 规则的名称
	UserId         int           `json:"userId,omitempty"`           // 规则的所有者(用户)的ID
	UserName       string        `json:"userName,omitempty"`         // 规则的所有者(用户)的用户名
	KeyWord        string        `json:"keyword,omitempty"`          // 报警关键字
	KeywordIndex   int           `json:"keywordIndex"`     // 报警关键字在日志中的索引，日志按空格分割后从0记录索引，报警关键字所在的索引
	CreateTime     int64         `json:"createTime,omitempty"`       // 报警规则的创建时间戳
	UpdateTime     int64         `json:"updateTime,omitempty"`       // 报警规则的更新时间戳
}

// 查询单条rule传入参数
type RuleId struct {
Id             int           `json:"ruleId,omitempty"`           // 报警规则的ID，每插入一条自增
}
// 更新rule传入参数
type RuleUpdate struct {
	Id             int           `json:"ruleId,omitempty"`           // 报警规则的ID，每插入一条自增
	KeyWord        string        `json:"keyword,omitempty"`          // 报警关键字
	KeywordIndex   int           `json:"keywordIndex"`     // 报警关键字在日志中的索引，日志按空格分割后从0记录索引，报警关键字所在的索引
}

/* 日志报警服务维护的报警任务表 */
type Task struct {
	Id                int         `json:"taskId"`          // 报警任务在报警任务表中的ID，每插入一条自增
	TaskName          string      `json:"taskName"`        // 报警任务的名称
	RuleId            int         `json:"ruleId"`          // 报警规则ID
	RuleName          string      `json:"ruleName"`        // 报警规则名称
	KeyWord           string      `json:"keyword"`         // 报警关键字
	KeywordIndex      int         `json:"keywordIndex"`    // 报警关键字在日志中的索引，日志按空格分割后从0记录索引，报警关键字所在的索引
	TimeWindow        int         `json:"timeWindow"`      // 报警任务的监控周期
	ThresholdNum      int         `json:"thresholdNum"`    // 报警任务的报警日志条数阈值，在监控周期内一旦达到阈值，立即报警
	AlarmGroupID      int         `json:"alarmGroupId"`    // 报警用户组ID
	AlarmGroupName    string      `json:"alarmGroupName"`  // 报警用户组名称
	FilePath         string         `json:"filePath"`         // 报警任务所属的文件的路径名
	FileName         string         `json:"fileName"`         // 报警任务所属的文件的文件名
	FileId           int              `json:"fileId"`          // 报警任务所属的文件在文件表中的ID
	ContainerId          int              `json:"containerId"`          // 报警服务所属ContainerID
	ServiceId          int              `json:"serviceId"`          // 报警任务所属服务ID
	RoomId          int              `json:"roomId"`          // 报警任务所属Room的ID
	UserId          int              `json:"userId"`          // 报警任务所属User的ID
	CreateTime        int64       `json:"createTime"`      // 报警任务的创建时间戳
	UpdateTime        int64       `json:"updateTime"`      // 报警任务的更新时间戳
	TaskState         string      `json:"taskState"`       // 报警任务的当前运行状态
}

/* 日志报警服务维护的 日志文件表 */
type File struct {
	Id            int           `json:"fileId"`           // 报警日志文件在日志文件表中的ID，每插入一条自增
	FileName      string        `json:"fileName"`         // 报警文件名称
	FilePath      string        `json:"filePath"`         // 报警文件的路径名
	ContainerId   int           `json:"containerId,omitempty"`      // 报警文件所属的容器在容器表中的ID
	ServiceId   int             `json:"serviceId,omitempty"`      // 报警文件所属服务的ID
	CreateTime    int64         `json:"createTime"`       // 创建时间戳
	Tasks         []Task        `json:"tasks"`
}

/* 日志报警服务维护的 Container 表 */
type Container struct {
	Id            int           `json:"containerId"`      // 容器的 ID，来自页面调用接口传入
	ContainerName string        `json:"containerName"`    // 容器的名称
	ServiceId     int           `json:"sercviceId"`       // 报警容器所属的服务
	CreateTime    int64         `json:"createTime"`       // 创建时间戳
	Files         []File        `json:"files"`
}

/* 日志报警服务维护的 Service 表 */
type Service struct {
	Id            int           `json:"serviceId"`       // 服务 ID，来自页面调用接口传入
	ServiceName   string        `json:"serviceName"`      // 服务的名称
	TaskNameByUser string       `json:"taskName"`        // 用户指定的服务为单位提交的任务名称
	RoomId        int           `json:"roomId"`           // 报警服务所属的 Room 的 ID (Room表中)
	CreateTime    int64         `json:"createTime"`       // 创建时间戳
	Containers    []Container   `json:"containers"`
}

/* 日志报警服务维护的 room 表 */
type 	Room struct {
	Id            int           `json:"roomId"`           // Room ID，来自页面调用接口传入
	RoomName      string        `json:"roomName"`         // Room 的名称
	UserId        int           `json:"userId"`           // 页面传入的用户ID
	CreateTime    int64         `json:"createTime"`       // 报警任务的创建时间戳
	ServiceInfo   Service       `json:"service"`
}

/* 日志报警服务维护的 User 表 */
type User struct {
	Id            int           `json:"userId"`           // 页面传入的用户ID
	UserName      string        `json:"userName"`         // Room 的名称
	CreateTime    int64         `json:"createTime"`       // 报警任务的创建时间戳
}

/* 日志报警服务维护的 用户自已创建的服务任务名 表(用户会以服务为单位指定一个任务名) */
type TaskNameByUser struct {
	UserId      int             `json:"userId"`         // 用户ID
	ServiceId   int             `json:"serviceId"`       // 用户ID
	TaskNameByUser string       `json:"taskName"`        // 用户指定的服务为单位提交的任务名称
}



/* 以服务为单位提交任务的信息 */
type SubmitTasks struct {
	UserId     int           `json:"userId"`          // 用户（整个报警任务的提交者，也是登录的用户）ID
	UserName   string        `json:"userName"`        // 用户名称
	TaskNameByUser string    `json:"taskName"`        // 用户指定的服务为单位提交的任务名称
	RoomInfo   Room          `json:"room"`
}


// 6. HTTP 理由信息
/* HTTP 路由信息 */
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}


// 7. 按任务ID 删除任务的请求体
type DeleteTask struct {
	TaskID   int          `json:"taskId"`
}

// 8. 获取用户已创建报警任务的服务列表的请求体
/* 报警任务的服务列表的查询请求 */
type ListServiceRequest struct{
	UserId          int      `json:"userId"`
	UserName        string   `json:"userName"`
}

type ServicesInfo struct{
	ServiceId          int      `json:"serviceId,omitempty"`
	ServiceName        string   `json:"serviceName,omitempty"`
}

// 以服务为单位删除任务时提供的serviceId
type ServiceID struct{
	ServiceId          int      `json:"serviceId,omitempty"`
}

// 用于暂停/重新启动报警任务接口，传入的是底层（storm 上）的任务名称列表
type StopStartTask struct{
	UserId		int 	     `json:"userId"`
	TaskNames	[]TaskName   `json:"taskNames"`
}
// 底层（storm 上）的任务名称
type TaskName struct {
	TaskName	string  `json:"taskName"`
}
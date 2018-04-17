package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/types"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/utils"
	"github.com/golang/glog"
	"net/url"
	"fmt"
	"time"
	"strconv"
)

/* 连接Mysql数据库 */
/*
func ConnectMYSQL() (*sql.DB, error) {
	/*DSN数据源名称
	  [username[:password]@][protocol[(address)]]/dbname[?param1=value1¶mN=valueN]
	  user@unix(/path/to/socket)/dbname
	  user:password@tcp(localhost:5555)/dbname?charset=utf8&autocommit=true
	  user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?charset=utf8mb4,utf8
	  user:password@/dbname
	  无数据库: user:password@/
	*/
/*	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/?charset=utf8") //第一个参数为驱动名
	if err != nil {
		return nil, err
	}
	return db, nil
}
*/
func ConnectMYSQL()  (*sql.DB, error) {
	var c utils.Conf
	var err error
	_, err=c.GetConf()
	if err !=nil {
		glog.Error(err)
		glog.Error("Failed to connect Mysql due to can not read connect configuration")
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8&loc=%s&parseTime=true", c.MysqlUser, c.MysqlPassword, c.MysqlHost, c.MysqlPort, url.QueryEscape("Asia/Shanghai"))
	db, err := sql.Open("mysql", uri)
	if err != nil {
		glog.Error(err)
		glog.Error("Failed to onnect Mysql due to open Mysql")
		return nil, err
	}

	// 连接Mysql成功时确保数据库已经成功创建，如果未成功创建，则接下来的表操作也无法进行
	ret, err:= CreteDatabase(db, types.DBname); if !ret{
		glog.Error(err)
		glog.Errorf("Falied to verify the existences database %s in Mysql", types.DBname)
		return nil, err
	}
	return  db, nil
}

/* 创建用户数据库 */
func CreteDatabase(db *sql.DB, dbname string) (bool, error) {

	//db.Query("SET GLOBAL sql_mode = '';")
	rows, err := db.Query("create database if not exists "  + dbname)
	if err != nil{
		glog.Error(err)
		glog.Errorf("Create database failed: %v", err)
		return false, err
	}
	defer rows.Close()
	return true, nil
}

/* 创建表 */
func CreteTable(db *sql.DB, tablename string, createcmd string) bool {

	//Rule表{ruleid,rulename,username,userid,keyword,keywordindex,createtime,updatetime}
	//_, err := db.Query("create table if not exists  "+ dbname + ".rules(ruleid int auto_increment, rulename varchar(255), userid int, username varchar(255), keyword varchar(255), keywordindex int, createtime bigint, updatetime bigint, primary key(ruleid));")
	rows, err := db.Query(createcmd)
	defer rows.Close()
	if err != nil{
		glog.Error(err)
		glog.Errorf("Create table %s failed", tablename)
		return false
	}
	glog.Infof("Create table %s successed", tablename)
	return true
}


// 插入数据时，如果在检查数据存在性过程中失败，则默认有记录存在，返回插入失败，以避免数据被恶意或者误操作修改/更性
func InsertTableData(db *sql.DB, tx *sql.Tx,  tabname string, input interface{}) bool {

	// 1. Rule 表插入数据
	// Rule表{ id | rulename | userid | username | keyword | keywordindex | createtimestamp | updatetimestamp }
	// 插入值(rulename,userid,username,keyword,keywordindex,createtimestamp,updatetimestampe)
	if val, ok := input.(types.Rule);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix() *1000
		if tabname != types.RuleTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.RuleTable)
			return false
		}
		//sqlstr :=  "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.RuleDataExistedCMD + "userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		if RecordExisted(db, types.RuleTable, sqlCMD) {
			glog.Errorf("Existed record for [userid=%d rulename=%s] in table %s, refuse to insert new one, try update it.\n", val.UserId, val.RuleName, types.RuleTable)
			return false
		}

		//tx.Exec("INSERT INTO user(uid,username,age) values(?,?,?)",i,"user"+strconv.Itoa(i),i-1000)
		_, err := tx.Exec(types.RuleTableInsertElements, val.RuleName, val.UserId, val.UserName, val.KeyWord, val.KeywordIndex, timestamp, timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.RuleTable)
			return false
		}

		return true
	}

	// 2. User 表插入数据
	// User表{ uid | username | createtimestamp }
	// 插入值(uid,username,createtimestamp)
	if val, ok := input.(types.User);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.UserTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.UserTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.UserDataExistedCMD  + strconv.Itoa(val.Id)  + " limit 1"
		if RecordExisted(db, types.UserTable, sqlCMD) {
			glog.Errorf("Existed record for [userid=%d] in table %s, refuse to insert new one, try update it.\n", val.Id, types.UserTable)
			return true
		}

		_, err := tx.Exec(types.UserTableInsertElements, val.Id, val.UserName,timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.UserTable)
			return false
		}

		return true
	}

	// 3. Room 表插入数据
	// Room表{ id | roomname  | uid  | createtimestamp  }
	// 插入值(id,roomname,uid,createtimestamp)
	if val, ok := input.(types.Room);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.RoomTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.RoomTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.RoomDataExistedCMD + strconv.Itoa(val.Id) + " limit 1"
		if RecordExisted(db, types.RoomTable, sqlCMD) {
			glog.Errorf("Existed record for [roomid=%d] in table %s, refuse to insert new one, try update it.\n", val.Id, types.RoomTable)
			return true
		}

		_, err := tx.Exec(types.RoomTableInsertElements, val.Id, val.RoomName, val.UserId, timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.RoomTable)
			return false
		}

		return true
	}

	// 4. Service 表插入数据
	// Service表{ id | servicename | rid  | createtimestamp }
	// 插入值(id,servicename,rid,createtimestamp)
	if val, ok := input.(types.Service);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.ServiceTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.ServiceTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.ServiceDataExistedCMD + strconv.Itoa(val.Id) + " limit 1"
		if RecordExisted(db, types.ServiceTable, sqlCMD) {
			glog.Errorf("Existed record for [serviceid=%d] in table %s, refuse to insert new one, try update it.\n", val.Id, types.ServiceTable)
			return true
		}

		_, err := tx.Exec(types.ServiceTableInsertElements, val.Id, val.ServiceName, val.RoomId,val.TaskNameByUser, timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.ServiceTable)
			return false
		}

		return true
	}

	// 5. Container 表插入数据
	// Container表{ id | containername | sid  | createtimestamp }
	// 插入值(id | containername | sid  | createtimestamp)
	if val, ok := input.(types.Container);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.ContainerTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.ContainerTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.ContainerDataExistedCMD + strconv.Itoa(val.Id) + " limit 1"
		if RecordExisted(db, types.ContainerTable, sqlCMD) {
			glog.Errorf("Existed record for [containerid=%d] in table %s, refuse to insert new one, try update it.\n", val.Id, types.ContainerTable)
			return true
		}

		_, err := tx.Exec(types.ContainerTableInsertElements, val.Id, val.ContainerName, val.ServiceId, timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.ContainerTable)
			return false
		}

		return true
	}

	// 6. File 表插入数据
	// File表{ id | filename  | filepath  | cid  | createtimestamp }
	// 插入值(filename,filepath,cid,createtimestamp)
	if val, ok := input.(types.File);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.FileTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.FileTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.FileDataExistedCMD + "filename='" + val.FileName + "' and filepath='" + val.FilePath + "' and cid=" + strconv.Itoa(val.ContainerId) + " and sid=" + strconv.Itoa(val.ServiceId) + " limit 1"

		//fmt.Printf("SQL cmd : %s\n", sqlCMD)
		if RecordExisted(db, types.FileTable, sqlCMD) {
			glog.Errorf("Existed record for [filename=%s filepath=%s containerid=%d serviceid=%d] in table %s, refuse to insert new one, try update it.\n", val.FileName, val.FilePath, val.ContainerId, val.ServiceId, types.FileTable)
			return true
		}

		_, err := tx.Exec(types.FileTableInsertElements, val.FileName, val.FilePath, val.ContainerId, val.ServiceId,  timestamp)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.FileTable)
			return false
		}
		return true
	}

	// 7. TaskNameByUser 表插入数据
	// TaskNameByUser表{ id | username | taskname }
	// 插入值(id,username,taskname)
	if val, ok := input.(types.TaskNameByUser);ok{
		//fmt.Println(val)
		if tabname != types.TaskNameTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.TaskNameTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.TaskNameTable + " where username=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.TaskNameDataExistedCMD  + strconv.Itoa(val.UserId)  + " and serviceid=" + strconv.Itoa(val.ServiceId) + " and taskname='" + val.TaskNameByUser + "' limit 1"
		if RecordExisted(db, types.UserTable, sqlCMD) {
			glog.Errorf("Existed record for [userid=%d, serviceid=%d, taskname=%s] in table %s, refuse to insert new one, try update it.\n", val.UserId, val.ServiceId, val.TaskNameByUser, types.TaskNameTable)
			return true
		}

		_, err := tx.Exec(types.TaskNameTableInsertElements, val.UserId, val.ServiceId, val.TaskNameByUser)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.TaskNameTable)
			return false
		}

		return true
	}

	// 8. Task 表插入数据
	// Task表{ id | taskname | ruleid | rulename | keyword | keywordindex | timewindow | number | groupid | groupname | fid  | createtimestamp | updatetimestamp }
	// 插入值(taskname,ruleid,rulename,keyword,keywordindex,timewindow,number,groupid,groupname,fid,createtimestamp,updatetimestamp)
	if val, ok := input.(types.Task);ok{
		//fmt.Println(val)
		timestamp := time.Now().Unix()*1000
		if tabname != types.TaskTable {
			glog.Errorf("Target table %s not matched the request table %s\n", tabname, types.TaskTable)
			return false
		}
		//sqlstr := "select count(*) from " + dbname + "." + types.RuleTable + " where userid=" + strconv.Itoa(val.UserId) + " and rulename='"+ val.RuleName+"' limit 1"
		sqlCMD := types.TaskDataExistedCMD + val.TaskName + "'  limit 1"
		if RecordExisted(db, types.TaskTable, sqlCMD) {
			glog.Errorf("Existed record for [taskname=%s] in table %s, refuse to insert new one, try update it.\n", val.TaskName, types.TaskTable)
			return false
		}

		_, err := tx.Exec(types.TaskTableInsertElements, val.TaskName, val.RuleId, val.RuleName, val.KeyWord, val.KeywordIndex, val.TimeWindow, val.ThresholdNum, val.AlarmGroupID , val.AlarmGroupName, val.FilePath, val.FileName, val.FileId, val.ServiceId, val.ContainerId, val.RoomId, val.UserId,timestamp, timestamp,val.TaskState)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Insert record into table %s failed\n", types.TaskTable)
			return false
		}

		return true
	}else{
		glog.Error("Failed to do insert, no matched table found")
		return false
	}
}


/* 日志报警规则表中更新指定用户指定规则的数据 */
func Update_Rule(db *sql.DB, dbname string, rule types.Rule ) bool {

	stmt, err := db.Prepare("update " + dbname + ".rules set keyword=?,keywordindex=?,updatetimestamp=? where id=?")
	if err != nil{
		glog.Error(err)
		glog.Errorf("Preprae to update rules [ruleId=%d, rulename=%s] failed", rule.Id, rule.RuleName)
		return false
	}

	res, err := stmt.Exec(rule.KeyWord, rule.KeywordIndex, rule.UpdateTime, rule.Id)
	if err != nil{
		glog.Error(err)
		glog.Errorf("Update rules [ruleId=%d] in table rules failed", rule.Id)
		return false
	}

	affect, err := res.RowsAffected() //RowsAffected returns the number of rows affected by an update, insert, or delete.
	glog.Infof("Affected rows: %d", affect)
	glog.Infof("Update rule [ruleId=%d] successed", rule.Id)
	return true
}


/* 暂停/启动日志报警任务时，更新日志报警任务表中任务的运行状态数据 */
func Update_TaskStatus(tx *sql.Tx, dbname string, tabname string, taskname string , taskstate string) bool {
	/*
	stmt, err := db.Prepare("update " + dbname + ".tasks set taskstate=? where taskname=?")
	if err != nil{
		glog.Error(err)
		glog.Errorf("Preprae to update tasks [taskname=%s, taskstate=%s] failed", taskname, taskstate)
		return false
	}

	res, err := stmt.Exec(taskstate, taskname)
	if err != nil{
		glog.Error(err)
		glog.Errorf("Update tasks [taskname=%s, taskstate=%s] failed", taskname, taskstate)
		return false
	}

	affect, err := res.RowsAffected() //RowsAffected returns the number of rows affected by an update, insert, or delete.
	glog.Infof("Affected rows: %d", affect)
	glog.Infof("Update tasks [taskname=%s, taskstate=%s] successed", taskname, taskstate)
	return true
	*/

	_, err := tx.Exec("update " + dbname + "." + tabname + " set taskstate=? where taskname=?", taskstate,taskname)
	if err != nil {
		glog.Error(err)
		return false
	}
	glog.Infof("Update tasks [taskname=%s, taskstate=%s] successed", taskname, taskstate)
	return true
}


/* 查找日志报警规则表 */
func SelectRules(db *sql.DB, dbname string, tabname string, req types.ListRequest) ([]types.Rule, error) {


	offset := (req.CurrentPage - 1) * req.PageSize
	if req.UserId != 0 {      // 查询指定用户的规则
		rows, err := db.Query("select * from "+ dbname + "." + tabname + "  where userid=" + strconv.Itoa(req.UserId) + " limit " + strconv.Itoa(offset) + "," + strconv.Itoa(req.PageSize))
		if err != nil {
			glog.Error(err)
			glog.Errorf("Select data from table %s for userid=%d failed", tabname, req.UserId)
			return nil, err
		}
		defer rows.Close()
		return Return_Rules(rows), nil
	} else {     // 查询所有用户的规则
		rows, err := db.Query("select * from " + dbname + "." + tabname + " limit " + strconv.Itoa(offset) + "," + strconv.Itoa(req.PageSize))
		if err != nil {
			glog.Error(err)
			glog.Errorf("Select data from %s failed", tabname)
			return nil, err
		}
		defer rows.Close()
		return Return_Rules(rows), nil
	}
}



func SelectUsers(db *sql.DB, dbname string, tabname string) ([]types.RetUser, error) {

	var users []types.RetUser
	rows, err := db.Query("select id,username from "+ dbname + "." + tabname)
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from table %s failed", tabname)
		return users, err
	}
	defer rows.Close()


	for rows.Next() {
		var UserId int
		var UserName string

		err := rows.Scan(&UserId, &UserName)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Try to get data failed in scan data in table %s. Error info: %s \n", tabname, err)
			return users,err
		}
		temp:=types.RetUser{ID: UserId, UserName: UserName,}
		users=append(users, temp)
	}
	return users,nil
}

func SelectRooms(db *sql.DB, dbname string, tabname string, userId int) ([]types.RetRoom, error) {

	var rooms []types.RetRoom
	rows, err := db.Query("select id,roomname from "+ dbname + "." + tabname + " where uid=" + strconv.Itoa(userId))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return rooms, err
	}
	defer rows.Close()


	for rows.Next() {
		var RoomId int
		var RoomName string

		err := rows.Scan(&RoomId, &RoomName)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Try to get data failed in scan data in table %s. Error info: %s \n", tabname, err)
			return rooms, err
		}
		temp:=types.RetRoom{ID: RoomId, RoomName: RoomName,}
		rooms=append(rooms, temp)
	}
	return rooms,nil
}

func SelectServices(db *sql.DB, dbname string, tabname string, roomId int) ([]types.RetService, error) {

	var services []types.RetService
	rows, err := db.Query("select id,servicename,taskname from "+ dbname + "." + tabname + " where rid=" + strconv.Itoa(roomId))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return services, err
	}
	defer rows.Close()


	for rows.Next() {
		var ServiceId int
		var ServiceName string
		var TaskNameByUser string

		err := rows.Scan(&ServiceId, &ServiceName, &TaskNameByUser)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Try to get data failed in scan data in table %s. Error info: %s \n", tabname, err)
			return services, err
		}
		temp:=types.RetService{ID: ServiceId, ServiceName: ServiceName, TaskNameByUser: TaskNameByUser}
		services=append(services, temp)
	}
	return services,nil
}


func SelectContainers(db *sql.DB, dbname string, tabname string, serviceId int) ([]types.RetContainer, error) {

	var containers []types.RetContainer
	rows, err := db.Query("select id,containername from "+ dbname + "." + tabname + " where sid=" + strconv.Itoa(serviceId))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return containers, err
	}
	defer rows.Close()


	for rows.Next() {
		var ContainerId int
		var ContainerName string

		err := rows.Scan(&ContainerId, &ContainerName)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Try to get data failed in scan data in table %s. Error info: %s \n", tabname, err)
			return containers, err
		}
		temp:=types.RetContainer{ID: ContainerId, ContainerName: ContainerName,}
		containers=append(containers, temp)
	}
	return containers,nil
}

func SelectFiles(db *sql.DB, dbname string, tabname string, containerId int) ([]types.RetFile, error) {

	var files []types.RetFile
	rows, err := db.Query("select id,filename,filepath,cid,sid from "+ dbname + "." + tabname + " where cid=" + strconv.Itoa(containerId))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return files, err
	}
	defer rows.Close()


	for rows.Next() {
		var FileId int
		var FileName string
		var FilePath string
		var ContainerId int
		var ServiceId  int

		err := rows.Scan(&FileId, &FileName, &FilePath, &ContainerId, &ServiceId)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Try to get data failed in scan data in table %s. Error info: %s \n", tabname, err)
			return files, err
		}
		temp:=types.RetFile{Id: FileId, FileName: FileName, FilePath: FilePath, ContainerId: ContainerId, ServiceId: ServiceId}
		files=append(files, temp)
	}
	return files,nil
}

func SelectFileTasks(db *sql.DB, dbname string, tabname string, fileId int /*, req types.RuleSearchRequest*/) ([]types.Task, error) {

	var tasks []types.Task

	// offset := (req.CurrentPage - 1) * req.PageSize
	//rows, err := db.Query("select * from " + dbname + "." + tabname + " where fileid=" + strconv.Itoa(fileId) + " limit " + strconv.Itoa(offset) + "," + strconv.Itoa(req.PageSize))

	rows, err := db.Query("select * from " + dbname + "." + tabname + " where fileid=" + strconv.Itoa(fileId))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return tasks, err
	}
	defer rows.Close()
	return Return_Tasks(rows), nil
}

func SelectTasksByTaskName(db *sql.DB, dbname string, tabname string, taskname string /*, req types.RuleSearchRequest*/) ([]types.Task, error) {

	var tasks []types.Task
	rows, err := db.Query("select * from " + dbname + "." + tabname + " where taskname='" + taskname + "'")
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select data from %s failed", tabname)
		return tasks, err
	}
	defer rows.Close()
	return Return_Tasks(rows), nil
}

// 统计指定表中属于特定用户的数据记录数量，不指定用户则是整表全量统计
func GetCount(db *sql.DB, dbname string, tabname string, req types.ListRequest) (bool, int) {
	if req.UserId != 0 {      // 查询指定用户
		rows, err := db.Query("select count(*) from "+ dbname + "." + tabname + " where userid=" + strconv.Itoa(req.UserId))
		if err != nil {
			glog.Error(err)
			glog.Errorf("Get count from table %s for userid=%d failed",tabname, req.UserId)
			return false, 0
		}
		defer rows.Close()
		rows.Next()
		var count int
		err = rows.Scan(&count)
		if err != nil {
			glog.Error(err)
			return  false, 0
		}
		return true, count
	} else {     // 查询所有用户
		rows, err := db.Query("select count(*) from " + dbname + "." + tabname)
		if err != nil {
			glog.Error(err)
			glog.Errorf("Get the count from table %s failed", tabname)
			return false, 0
		}
		defer rows.Close()
		rows.Next()
		var count int
		err = rows.Scan(&count)
		if err != nil {
			glog.Error(err)
			return   false, 0
		}
		return true, count
	}
}

/* 查看指定规则ID的日志报警规则 */
func SelectRule(db *sql.DB, dbname string, tabname string, id int)  (types.Rule, error) {

	rows, err := db.Query("select * from "+ dbname + "." + tabname + " where id=" + strconv.Itoa(id))
	if err != nil {
		glog.Error(err)
		glog.Errorf("Select rules for ruleid=%d failed", id)
		return types.Rule{}, err
	}

	defer rows.Close()
	ret := Return_Rules(rows)
	if len(ret) !=1 {
		glog.Errorf("No matched record found for ruleid=%d  in table %s", id, tabname)
		return types.Rule{}, err
	}
	rule := ret[0]
	return rule, nil
}

/* 查看指定服务ID的报警任务(用户创建) */
func SelectTaskNameByUser(db *sql.DB, dbname string,  tabname string, serviceid int)  (types.TaskNameByUser, error) {

	var ret types.TaskNameByUser
	err := db.QueryRow("select userid,taskname from "+ dbname + "." + tabname + " where serviceid=" + strconv.Itoa(serviceid)).Scan(&ret.UserId, &ret.TaskNameByUser)
	if err != nil {
		glog.Error(err)
		return ret, err
	}
	ret.ServiceId = serviceid

	return ret, nil
}


/* 通过ID删除指定表中的数据 */
func DeleteByID(db *sql.DB, dbname string,tabname string, id int) bool {

	stmt, err := db.Prepare("delete from " + dbname + "." + tabname + " where id=?")
	if err != nil {
		glog.Error(err)
		glog.Errorf("Prepare to delete table %s record [id=%d] failed", tabname, id)
		return false
	}

	res, err := stmt.Exec(id)
	if err != nil {
		glog.Error(err)
		glog.Errorf("Delete table %s record  [id=%d] failed", tabname, id)
		return false
	}

	affect, err := res.RowsAffected()
	glog.Infof("Deleted rows: %d", affect)
	if affect > 0 {
		return true
	}else {
		glog.Errorf("No record in  table %s for  [id=%d] to delete", tabname, id)
		return false
	}
}



// 检查表总是否有某条记录，如果检查过程中失败，默认有记录存在，以避免数据被恶意或者误操作修改/更性
func RecordExisted(db *sql.DB, tabname string, cmd string) bool {

	count :=1
	//rows, err := db.Query("select count(*) from " + dbname + "." + tabname+ " where userid=" + strconv.Itoa(req.UserId) + " limit 1")
	rows, err := db.Query(cmd)
	if err != nil {
		glog.Error(err)
		glog.Errorf("Check the record existence in table %s failed when query", tabname)
		return true
	}
	defer rows.Close()

	if !rows.Next() {
		glog.Errorf("Check the record existence in table %s failed when count", tabname)
		return true
	}

	err = rows.Scan(&count)
	if err != nil || count == 1 {
		glog.Error(err)
		return true
	}

	return false
}



func Return_Rules(rows *sql.Rows)  []types.Rule{
	var tmp []types.Rule
	for rows.Next() {
		var Id int
		var UserId int
		var UserName string
		var RuleName string
		var KeyWord string
		var KeywordIndex int
		var CreateTime int64
		var UpdateTime int64

		err := rows.Scan(&Id, &RuleName, &UserId, &UserName, &KeyWord, &KeywordIndex, &CreateTime, &UpdateTime)
		if err != nil {
			glog.Error(err)
			glog.Warningf("Try to get rules failed in scan data. Error info: %s. ", err)
			panic(err)
		}
		temp:=types.Rule{Id: Id, UserName: UserName, UserId: UserId, RuleName: RuleName, KeyWord: KeyWord, KeywordIndex: KeywordIndex, CreateTime: CreateTime, UpdateTime: UpdateTime}
		tmp=append(tmp, temp)
	}
	return tmp
}



func Return_Tasks(rows *sql.Rows)  []types.Task{
	var tmp []types.Task
	for rows.Next() {
		var TaskId int
		var Taskname string
		var RuleId   int
		var RuleName string
		var KeyWord string
		var KeywordIndex int
		var TimeWindow  int
		var ThresholdNum int
		var AlarmGroupID      int
		var AlarmGroupName    string
		var FilePath   string
		var FileName   string
		var FileId   int
		var ContainerId   int
		var ServiceId  int
		var RoomId   int
		var UserId   int
		var CreateTime int64
		var UpdateTime int64
		var TaskState string


		err := rows.Scan(&TaskId, &Taskname, &RuleId, &RuleName, &KeyWord, &KeywordIndex, &TimeWindow, &ThresholdNum, &AlarmGroupID,
			&AlarmGroupName, &FilePath, &FileName, &FileId, &ContainerId, &ServiceId, &RoomId, &UserId, &CreateTime, &UpdateTime, &TaskState)
		if err != nil {
			glog.Error(err)
			glog.Warningf("Try to get tasks failed in scan data. Error info: %s. ", err)
			return tmp
		}
		temp:=types.Task{Id: TaskId, TaskName: Taskname, RuleId: RuleId, RuleName: RuleName, KeyWord: KeyWord, KeywordIndex: KeywordIndex,TimeWindow: TimeWindow,
			ThresholdNum: ThresholdNum, AlarmGroupID: AlarmGroupID, AlarmGroupName: AlarmGroupName, FilePath: FilePath, FileName:FileName, FileId: FileId, ContainerId: ContainerId,
			ServiceId: ServiceId, RoomId: RoomId, UserId: UserId, CreateTime: CreateTime, UpdateTime: UpdateTime, TaskState:TaskState}
		tmp=append(tmp, temp)
	}
	return tmp
}



func DeleteServiceRalatedData(db *sql.DB, dbname string, serviceId int) bool{
	//删除相关任务数据
	rows, _:=db.Query("select taskname from " + dbname  + "." + types.TaskTable + " where serviceid=?", serviceId)
	defer  rows.Close()
	if !rows.Next(){
		glog.Errorf("No tasks found in [serviceid=%d]\n",serviceId)
		//return false
	}else{
		var taskName string
		if err := rows.Scan(&taskName); err != nil {
			glog.Error(err)
			glog.Error("Failed to get tasks in this service")
		//	return false
		}

		StringCMD := "/opt/storm/bin/storm kill " + taskName
		if !utils.TopologySubmit(StringCMD, taskName) {
			glog.Errorf("Submit Topology task %s failed with CMD[%s]", taskName, StringCMD)
		//	return false
		}
	}
	for rows.Next() {
		var taskName string
		if err := rows.Scan(&taskName); err != nil {
			glog.Error(err)
			glog.Error("Failed to get tasks in this service")
		//	return false
		}

		StringCMD := "/opt/storm/bin/storm kill " + taskName
		if !utils.TopologySubmit(StringCMD, taskName) {
			glog.Errorf("Submit Topology task %s failed with CMD[%s]", taskName, StringCMD)
		//	return false
		}
	}

	tx, _:= db.Begin()
	defer tx.Commit()

	//删除相关任务数据
	_ ,err :=tx.Exec("delete from " +  dbname  + "." + types.TaskTable + " where serviceid=?", serviceId)
	if err != nil {
		glog.Error(err)
		glog.Error("Delete tasks data in table failed in DeleteServiceTasks()")
		//return false
	}

	//删除相关文件数据
	_ ,err =tx.Exec("delete from " +  dbname + "." + types.FileTable + " where sid=?", serviceId)
	if err != nil {
		glog.Error(err)
		glog.Error("Delete files data in table failed in DeleteServiceTasks()")
		//return false
	}


	//删除相关容器数据
	_ ,err =tx.Exec("delete from " +  dbname  + "." + types.ContainerTable + " where sid=?", serviceId)
	if err != nil {
		glog.Error(err)
		glog.Error("Delete containers data in table failed in DeleteServiceTasks()")
		//return false
	}

	//删除相关服务数据
	_ ,err =tx.Exec("delete from " +  dbname  + "." + types.ServiceTable + " where id=?", serviceId)
	if err != nil {
		glog.Error(err)
		glog.Error("Delete service data in table failed in DeleteServiceTasks()")
		//return false
	}

	//删除用户指定的以服务为单位的任务名数据
	_ ,err =tx.Exec("delete from " +  dbname  + "." + types.TaskNameTable + " where serviceid=?", serviceId)
	if err != nil {
		glog.Error(err)
		glog.Error("Delete service data in table failed in DeleteServiceTasks()")
		//return false
	}
	return true
}


func DB_Initial() bool {
	db, err := ConnectMYSQL()
	defer db.Close()
	if err != nil {
		glog.Error(err)
		glog.Error("Connect to Mysql failed in DB_Initial()")
		return false
	}

	tx, _ := db.Begin()
	defer tx.Commit()


	tableName := []tableCreate{{types.RuleTable,types.RuleTableCreateCMD},{types.TaskTable,types.TaskTableCreateCMD},{types.FileTable,types.FileTableCreateCMD}, {types.ContainerTable,types.ContainerTableCreateCMD},
		{types.ServiceTable,types.ServiceTableCreateCMD},{types.RoomTable,types.RoomTableCreateCMD},{types.UserTable,types.UserTableCreateCMD},{types.TaskNameTable,types.TaskNameTableCreateCMD}}

	for _, tc := range tableName{
		if ! CreteTable(db, tc.TabName, tc.TabCreateCMD) {
			glog.Error(err)
			glog.Errorf("Create table %s in database %s failed in DB_Initial()", tc.TabName,types.DBname)
			return false
		}
	}
	return true
}

type tableCreate struct{
	TabName string
	TabCreateCMD string
}

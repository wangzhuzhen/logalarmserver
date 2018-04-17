package commonCQL

import (
	"github.com/gocql/gocql"
	"strings"
	"encoding/json"
	"fmt"
	"log"
)

var (
	defaultSession *gocql.Session
)

func SetDefaultSession(s *gocql.Session) {
	defaultSession = s
}

type KeyspaceInterface interface {
	Name() 		string
	Session() 	*gocql.Session
}

type Keyspace struct {
	name 		string
	session 	*gocql.Session
}

func NewKeyspace(name string, session *gocql.Session) Keyspace{
	return Keyspace{
		name:name,
		session:session,
	}
}

func (ks Keyspace)Name() string{
	return ks.name
}

func (ks Keyspace)Session() *gocql.Session{
	return ks.session
}

func (ks Keyspace)CreateKeyspace(){
	if ks.session == nil{
		ks.session = defaultSession
	}
	replication := map[string]interface{}{
		"class": "SimpleStrategy",
		"replication_factor": 1,
	}
	replicationBytes, err := json.Marshal(replication)
	if err != nil{
		log.Fatalln("json marshal error")
	}
	replicationMap := strings.Replace(string(replicationBytes), `"`, `'`, -1)

	ks.session.Query(fmt.Sprintf("CREATE KEYSPACE %s WITH REPLICATION = %s",
		strings.Replace(string(ks.name), `"`, ``, -1), replicationMap)).Exec()
}

func (ks Keyspace)NewTable(name string, rowKeys, rangeKeys []string, row interface{}) Table{
	return Table{
		name:name,
		rowKeys:rowKeys,
		rangeKeys:rangeKeys,
		row:row,
		keyspace:ks,
		session:ks.session,
	}
}

func (ks Keyspace) CheckTableIsExist(name string) bool{
	err := ks.session.Query(fmt.Sprintf("DESCRIBE %s", name)).Exec()
	if err != nil{
		fmt.Println(fmt.Sprintf(" %s is exist in %s", name, ks.Name()))
		return false
	}
	return true;
}

func (ks Keyspace) DropKeyspace() error{
	return ks.session.Query(fmt.Sprintf("DROP KEYSPACE %s", strings.Replace(string(ks.name), `"`, ``, -1))).Exec()
}
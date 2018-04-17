package commonCQL

import (
	"github.com/gocql/gocql"
	"strings"
	"fmt"
	"reflect"
)

type tableInterface interface{
	Create() error
	Drop() error
	Query(statement string, params ...interface{}) Query
	Keyspace() Keyspace
	Name() string
	Row() interface{}
}

type Table struct {
	name 		string
	rowKeys 	[]string
	rangeKeys 	[]string
	row 		interface{}

	keyspace    Keyspace
	session 	*gocql.Session
}

func (t Table)Create() error{
	return t.create()
}

func (t Table)create(props ...string) error{
	rowKeys := t.rowKeys
	rangeKeys := t.rangeKeys

	pkString := "PRIMARY KEY (" + strings.Join(rowKeys, ", ")
	if len(rangeKeys) > 0 {
		pkString = pkString + "," + strings.Join(rangeKeys, ", ")
	}
	pkString = pkString + ")"

	fmt.Printf("pkString : ", pkString)

	fields := []string{}

	m, ok := StructToMap(t.Row())
	if !ok {
		fmt.Println("struct to map false")
		panic("Unable to get map from struct during create table")
	}

	fmt.Println("struct to map true %q", m)
	for key, value := range m{
		key = strings.ToLower(key)
		typ, err := stringTypeOf(value)
		if err != nil{
			return err
		}
		fields = append(fields, fmt.Sprintf(`%s %v`, key, typ))
	}

	fields = append(fields, pkString)

	propertiesString := ""
	if len(props) > 0 {
		propertiesString = "WITH " + strings.Join(props, " AND ")
	}
	fmt.Println("create table ", fmt.Sprintf(`CREATE TABLE %s.%s (%s) %s`, strings.Replace(string(t.keyspace.Name()), `"`, ``, -1), strings.Replace(string(t.name), `"`, ``, -1), strings.Join(fields, ", "), propertiesString))
	if t.session == nil{
		fmt.Println("t.session is nil")
	}
	err := t.session.Query(fmt.Sprintf(`CREATE TABLE %s.%s (%s) %s`, strings.Replace(string(t.keyspace.Name()), `"`, ``, -1), strings.Replace(string(t.name), `"`, ``, -1), strings.Join(fields, ", "), propertiesString)).Exec()

	return err
}

func (t Table) Drop() error {
	return t.session.Query(fmt.Sprintf(`DROP TABLE %q.%q`, t.Keyspace().Name(), t.Name())).Exec()
}

func (t Table) Query(statement string, values ...interface{}) Query {
	return Query{
		Statement: statement,
		Values:    values,

		Table:   t,
		Session: t.session,
	}
}

func (t Table) QueryToValues(statement string, values []interface{}) Query{
	return Query{
		Statement: statement,
		Values:    values,

		Table:   t,
		Session: t.session,
	}
}

func (t Table)Row() interface{}{
	return t.row
}

func (t Table) Keyspace() Keyspace {
	return t.keyspace
}
func (t Table) Name() string {
	return t.name
}

func (t Table) Insert(row interface{}){
	mapData, _ := StructToMap(row)
	strInsert := fmt.Sprintf(`INSERT INFO %s.%s (`, strings.Replace(string(t.keyspace.Name()), `"`, ``, -1), strings.Replace(string(t.name), `"`, ``, -1))
	var values = []interface{}{}
	index := 0
	for k, v := range mapData{
		if index != (len(mapData) - 1){
			strInsert += (k + ", ")
		}else {
			strInsert += k
		}
		index++
		values = append(values, v)
	}
	strInsert = strInsert + ") VALUES ("
	for i := 0; i < len(mapData); i++{
		if i != (len(mapData) - 1) {
			strInsert += "?, "
		}else{
			strInsert += "?"
		}
	}
	strInsert = strInsert + ")"
	err := t.QueryToValues("INSERT INTO ThrifeDemo.student (Email, Pwd) VALUES (?, ?)", values).Exec()
	fmt.Errorf("Insert error :",err)
}

func structToMap(obj interface{}) map[string]interface{}{
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	var  data = make(map[string] interface{})
	for i := 0; i < objType.NumField(); i++{
		data[objType.Field(i).Name] = objValue.Field(i).Interface()
	}
	return data
}
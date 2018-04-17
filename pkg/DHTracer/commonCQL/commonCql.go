package commonCQL

import (
	"github.com/gocql/gocql"
	"log"
)

func CreateClusterAndSession(address string) *gocql.Session{
	cluster := gocql.NewCluster(address)
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalln("Unable to open up a session with the Cassandra database (err=" + err.Error() + ")")
	}
	return session
}
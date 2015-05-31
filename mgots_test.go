package mgots

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"os"
	"testing"
	"time"
)

const (
	url          = "mongodb://localhost/mgots"
	testDb       = "mgots_test"
	testPageSize = 4096
	layout       = time.RFC3339Nano // "Jan 2, 2006 at 3:04pm (MST)"
)

func TestMain(m *testing.M) {
	// Connect to MongoDB and cleanup previous data
	database := DBConnect()
	database.DropDatabase()

	// Run the tests
	res := m.Run()

	// Cleanup
	//database.DropDatabase()

	// return
	os.Exit(res)
}

func DBConnect() *mgo.Database {
	session, err := mgo.Dial(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to %s with: %s", url, err.Error())
		os.Exit(1)
	}

	session.SetSafe(&mgo.Safe{W: 1})

	return session.DB(testDb)
}

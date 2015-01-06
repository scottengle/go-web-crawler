package main

import (
	"database/sql"

	"github.com/coopernurse/gorp"
)

func initDb() {
	dbmap := connect()

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via script
	err := dbmap.CreateTablesIfNotExists()
	Logger.checkErr(err, "Create tables failed")

	dbmap.TruncateTables()
	disconnect(dbmap)
}

func connect() *gorp.DbMap {
	Lock.Lock()

	// connect to db using standard Go database/sql API
	// use whatever database/sql driver you wish
	db, err := sql.Open("sqlite3", "post_db.bin")
	Logger.checkErr(err, "sql.Open failed")

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	// add a table, setting the table name to 'posts' and
	// specifying that the Id property is an auto incrementing PK
	dbmap.AddTableWithName(Link{}, "links").SetKeys(true, "ID")

	return dbmap
}

func disconnect(dbmap *gorp.DbMap) {
	dbmap.Db.Close()
	Lock.Unlock()
}

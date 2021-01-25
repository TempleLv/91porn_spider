package doneDB

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"spider91/catch"
)

type VideoDB struct {
	db *sql.DB
}

func (v *VideoDB) AddDone(vis []*catch.VideoInfo) (err error) {

	return
}

func (v *VideoDB) Close() error {
	return v.db.Close()
}

func OpenVDB(filename string) (vdb *VideoDB, err error) {
	vdb = new(VideoDB)
	vdb.db, err = sql.Open("sqlite3", filename)
	if err == nil {

		sql_table := `
		CREATE TABLE IF NOT EXISTS "done" (
			"viewkey" VARCHAR(64) PRIMARY KEY,
			"title" VARCHAR(64) NULL,
			"UpTime" TIMESTAMP default (datetime('now', 'localtime'))  
		);
		CREATE TABLE IF NOT EXISTS "undone" (
			"viewkey" VARCHAR(64) PRIMARY KEY,
			"title" VARCHAR(64) NULL,
			"UpTime" TIMESTAMP default (datetime('now', 'localtime')),
			"failcount" INTEGER NOT NULL DEFAULT 0
		);
	   `
		var res sql.Result
		res, err = vdb.db.Exec(sql_table)
		fmt.Println(res, err)
		return
	}

	return nil, err
}

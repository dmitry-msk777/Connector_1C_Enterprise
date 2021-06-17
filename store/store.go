package store

import (
	"database/sql"
	log "sqlquerybuilder/log"

	"sqlquerybuilder/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitBD() error {

	var err error
	DB, err = sql.Open("postgres", config.Conf.DatabaseURL)
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}
	return nil

}

func ExampleData() {
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS
		phonebook777("id" SERIAL PRIMARY KEY,
		"name" varchar(50), "phone" varchar(100))`)
	if err != nil {
		//fmt.Println(err.Error())
		log.Impl.Error(err.Error())
	}

	_, err = DB.Exec(`ALTER TABLE phonebook777
		ADD COLUMN address666 varchar(30)`)
	if err != nil {
		//fmt.Println(err.Error())
		log.Impl.Error(err.Error())
	}

	_, err = DB.Exec(`CREATE INDEX IDX_PHONEBOOK777_NAME 
		ON PHONEBOOK777 USING BTREE (NAME)`)
	if err != nil {
		//fmt.Println(err.Error())
		log.Impl.Error(err.Error())
	}

	// rename INDEX
	_, err = DB.Exec(`ALTER INDEX IDX_PHONEBOOK777_NAME RENAME TO IDX_PHONEBOOK666_NAME`)
	if err != nil {
		//fmt.Println(err.Error())
		log.Impl.Error(err.Error())
	}
}

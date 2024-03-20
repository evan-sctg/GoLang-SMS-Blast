package main

import (
	"database/sql"
	"fmt"
)
import _ "github.com/go-sql-driver/mysql"

var db *sql.DB

func dbConn() {
	var err error
	db, err = sql.Open(dbDriver, dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName)
	if err != nil {
		fmt.Println(err)
	}
}

func getBrodcastList(uuid BroadcastUUID) PhoneNumbers {
	var phoneNumbers PhoneNumbers

	var sql = ""
	sql = "SELECT `c`.`first_name`, `c`.`last_name`, `c`.`phone_164c`"
	sql += " FROM `groups_broadcasts` `gb`, `contacts` `c`, `groups_contacts` `gc`, `groups` `g`"
	sql += " WHERE `c`.`uuid` = `gc`.`contacts_uuid` AND `gc`.`groups_uuid` = `g`.`uuid` AND `g`.`uuid` = `gb`.`groups_uuid` AND `gb`.`broadcasts_uuid` = ?"

	rows, err := db.Query(sql, uuid.Uuid)
	if err != nil {
		fmt.Println(err)
	}

	defer rows.Close()
	for rows.Next() {
		var phone_164c string
		err := rows.Scan(&phone_164c)
		if err != nil {
			fmt.Println(err)
		}

		phoneNumbers.Numbers = append(phoneNumbers.Numbers, phone_164c)
	}

	return phoneNumbers
}

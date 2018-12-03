package server

import (
	"database/sql"
)

// CreateTables 在資料庫建立所需的表格
func CreateTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user (
			username varchar(64), 
			password char(41),
			PRIMARY KEY (username)
		)`,
		`CREATE TABLE IF NOT EXISTS login (
			username varchar(64),
			token char(36),
			PRIMARY KEY (username)
		)`,
		`CREATE TABLE IF NOT EXISTS invite (
			inviter varchar(64), 
			invitee varchar(64)
		)`,
		`CREATE TABLE IF NOT EXISTS friend (
			user1 varchar(64), 
			user2 varchar(64)
		)`,
		`CREATE TABLE IF NOT EXISTS post (
			author varchar(64), 
			message text
		)`,
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

// TruncateTables 清除表格中所有資料
func TruncateTables(db *sql.DB) error {
	queries := []string{
		"TRUNCATE TABLE user",
		"TRUNCATE TABLE login",
		"TRUNCATE TABLE invite",
		"TRUNCATE TABLE friend",
		"TRUNCATE TABLE post",
	}
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

package models

import (
	"database/sql"
	"fmt"
)

func NewDB(sqlUser, sqlPass string) error {
	conn := "%s:%s@tcp(127.0.0.1:3306)/"
	db, err := sql.Open("mysql", fmt.Sprintf(conn, sqlUser, sqlPass))
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS snippetbox")
	if err != nil {
		db.Close()
		return err
	}

	db.Close()

	db, err = sql.Open("mysql", fmt.Sprintf(conn+"snippetbox", sqlUser, sqlPass))
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE snippets (
			id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
			title VARCHAR(100) NOT NULL,
			content TEXT NOT NULL,
			created DATETIME NOT NULL,
			expires DATETIME NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		return err
	}

	_, err = db.Exec("CREATE INDEX idx_snippets_created ON snippets(created)")
	if err != nil {
		db.Close()
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			hashed_password CHAR(60) NOT NULL,
			created DATETIME NOT NULL
		);`)
	if err != nil {
		db.Close()
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE sessions (
			token CHAR(43) PRIMARY KEY,
			data BLOB NOT NULL,
			expiry TIMESTAMP(6) NOT NULL
		);`)
	if err != nil {
		db.Close()
		return err
	}

	_, err = db.Exec("CREATE INDEX sessions_expiry_idx ON sessions (expiry)")
	if err != nil {
		db.Close()
		return err
	}

	_, err = db.Exec("ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email)")
	if err != nil {
		db.Close()
		return err
	}

	defer db.Close()

	return nil
}

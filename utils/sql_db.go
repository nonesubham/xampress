package utils

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func buildDSN(cfg Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.Db_user, cfg.Db_pass, cfg.Db_host, cfg.Db_port)
}

func quoteIdent(ident string) string {
	safe := strings.ReplaceAll(ident, "`", "")
	return "`" + safe + "`"
}

// IsMySQLRunning checks if the configured MySQL server is currently accepting connections.
func IsMySQLRunning() bool {
	cfg := GetConf()
	db, err := sql.Open("mysql", buildDSN(cfg))
	if err != nil {
		return false
	}
	defer db.Close()
	db.SetConnMaxLifetime(2 * time.Second)
	return db.Ping() == nil
}

// connDB opens a connection to the configured MySQL server.
func connDB() *sql.DB {
	cfg := GetConf()
	db, err := sql.Open("mysql", buildDSN(cfg))
	Chk_error(err, "Error opening database connection, check your database credentials")
	return db
}

// CreateDBase creates a new database if it does not already exist.
func CreateDBase(dbName string) {
	db := connDB()
	defer db.Close()
	stus := db.Ping()
	Chk_error(stus, "Error connecting to MySQL server, check your database credentials")
	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + quoteIdent(dbName))
	Chk_error(err, "Unable to create database "+dbName)
}

// CreateDBaseOnHost creates a new database on a specific host and port
func CreateDBaseOnHost(dbName string, host string, port int) {
	cfg := GetConf()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/", cfg.Db_user, cfg.Db_pass, host, port)
	db, err := sql.Open("mysql", dsn)
	Chk_error(err, "Error opening database connection")
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + quoteIdent(dbName))
	Chk_error(err, "Unable to create database "+dbName)
}

// ChkDB checks whether a database with db_name already exists.
func ChkDB(db_name string) bool {
	db := connDB()
	defer db.Close()
	stus := db.Ping()
	Chk_error(stus, "Error connecting to MySQL server.")
	res, err := db.Query("SHOW DATABASES LIKE ?", db_name)
	if err != nil {
		return false
	}
	defer res.Close()
	return res.Next()
}

// DeleteDBase drops the specified database.
func DeleteDBase(db_name string) {
	db := connDB()
	defer db.Close()
	stus := db.Ping()
	Chk_error(stus, "Error connecting to MySQL server.")
	_, err := db.Exec("DROP DATABASE IF EXISTS " + quoteIdent(db_name))
	Chk_error(err, "Unable to delete database "+db_name)
}

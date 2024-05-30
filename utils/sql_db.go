package utils

import (
	"database/sql"


	_ "github.com/go-sql-driver/mysql"
)

//Connect to db
func connDB() *sql.DB {
	var csf_data Config = GetConf()
	var db_uri string = csf_data.Db_user + ":" + csf_data.Db_pass + "@/mysql"
	db, err := sql.Open("mysql", db_uri)
	Chk_error(err, "Error opening database connection")
	return db
}
//Create database
func CreateDBase(dbName string) {
	var db *sql.DB = connDB()
	defer db.Close() 
	stus := db.Ping()
	Chk_error(stus, "Encountering error while connecting to MYSQL server.")
	_, err := db.Exec("CREATE DATABASE " + dbName)
	Chk_error(err, "Unable to create databse")
}
//Check database already exists
func ChkDB(db_name string)bool{
	var db *sql.DB = connDB()
	defer db.Close() 
	stus := db.Ping()
	Chk_error(stus, "Encountering error while connecting to MYSQL server.")
	res,_ := db.Query("SHOW DATABASES LIKE '"+db_name+"';")
	if res.Next(){
		return true
	}else{
		return false
	}
}
//Delete database
func DeleteDBase(db_name string){
	var db *sql.DB = connDB()
	defer db.Close() 
	stus := db.Ping()
	Chk_error(stus, "Encountering error while connecting to MYSQL server.")
	_, err := db.Exec("DROP DATABASE " + db_name)
	Chk_error(err, "Unable to delete databse")
}

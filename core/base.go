package core

import (
	"time"

	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func checkDb(connectionName string) bool {
	_db := GetConnection(connectionName)
	if _db == nil {
		return false
	}
	if _db.Statement == nil {
		return false
	}
	sqlDB, err := _db.DB()
	if err != nil {
		fmt.Println("Database connection is lost:", err)
		return false
	} else {
		err := sqlDB.Ping()
		if err != nil {
			fmt.Println("Database connection is lost:", err)
			return false
		}
		fmt.Println("Database connection is still alive.")
	}
	return true

}

func InitDB(dsn string, connectionName string) (*gorm.DB, error) {
	_db := GetConnection(connectionName)
	if checkDb(connectionName) {
		return _db, nil
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	_db = db
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(50)
	sqlDB.SetMaxOpenConns(150)
	sqlDB.SetConnMaxLifetime(time.Second * 25)
	addConnect(connectionName, db)
	return db, nil
}
func CloseDb(connectionName string) {
	_db := GetConnection(connectionName)
	if checkDb(connectionName) {
		sqlDB, err := _db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

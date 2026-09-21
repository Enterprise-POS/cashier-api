package client

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	MODE        string = os.Getenv("MODE")
	DB_HOST     string = os.Getenv("DB_HOST")
	DB_USER     string = os.Getenv("DB_USER")
	DB_PASSWORD string = os.Getenv("DB_PASSWORD")
	DB_PORT     string = os.Getenv("DB_PORT")
	DB_TIMEZONE string = os.Getenv("DB_TIMEZONE")

	DEV_DB_HOST     string = os.Getenv("DEV_DB_HOST")
	DEV_DB_USER     string = os.Getenv("DEV_DB_USER")
	DEV_DB_PASSWORD string = os.Getenv("DEV_DB_PASSWORD")
	DEV_DB_PORT     string = os.Getenv("DEV_DB_PORT")
	DEV_DB_TIMEZONE string = os.Getenv("DEV_DB_TIMEZONE")
)

func CreateGormClient() *gorm.DB {
	var dialect string
	if MODE == "prod" {
		dialect = fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=require TimeZone=%s", DB_HOST, DB_USER, DB_PASSWORD, DB_PORT, DB_TIMEZONE)
	} else {
		dialect = fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=require TimeZone=%s", DEV_DB_HOST, DEV_DB_USER, DEV_DB_PASSWORD, DEV_DB_PORT, DEV_DB_TIMEZONE)
	}

	logMode := logger.Warn
	if MODE == "prod" {
		logMode = logger.Silent
	}
	db, err := gorm.Open(postgres.Open(dialect), &gorm.Config{
		Logger: logger.Default.LogMode(logMode),
	})

	if err != nil {
		panic(fmt.Sprintf("cannot initialize client: %s", err.Error()))
	}

	return db
}

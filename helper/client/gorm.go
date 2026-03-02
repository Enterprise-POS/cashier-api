package client

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB_HOST     string = os.Getenv("DB_HOST")
	DB_USER     string = os.Getenv("DB_USER")
	DB_PASSWORD string = os.Getenv("DB_PASSWORD")
	DB_PORT     string = os.Getenv("DB_PORT")
	DB_TIMEZONE string = os.Getenv("DB_TIMEZONE")
)

func CreateGormClient() *gorm.DB {
	dialect := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable TimeZone=%s", DB_HOST, DB_USER, DB_PASSWORD, DB_PORT, DB_TIMEZONE)
	db, err := gorm.Open(postgres.Open(dialect), &gorm.Config{})

	if err != nil {
		panic(fmt.Sprintf("cannot initialize client: %s", err.Error()))
	}

	return db
}

package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Example struct {
	ID   string `json:"id" gorm:"primarykey"`
	Name string `json:"name"`
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("develop.db"), &gorm.Config{})
	if err != nil {
		panic("could not connect to db")
	}

	// TODO: handle migrations as a script?
	db.AutoMigrate(&Example{})

	return db
}

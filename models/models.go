package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID            string `json:"id" gorm:"primarykey"`
	Email         string `json:"email" gorm:"unique"`
	EmailVerified bool   `json:"email_verified"`
	PasswordHash  string `json:"-"`
}

func ConnectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("develop.db"), &gorm.Config{})
	if err != nil {
		panic("could not connect to db")
	}

	// TODO: handle migrations as a script
	db.AutoMigrate(&User{})

	return db
}

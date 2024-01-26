package models

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID            string    `json:"id" gorm:"primarykey"`
	Email         string    `json:"email" gorm:"unique"`
	EmailVerified bool      `json:"email_verified"`
	PasswordHash  string    `json:"-"`
	Otp           string    `json:"-"`
	OtpExpiresAt  time.Time `json:"-"`
}

func (u *User) SendMail(message string) {
	fmt.Println("=====")
	fmt.Println("email to", u.Email)
	fmt.Println(message)
	fmt.Println("=====")
}

func (user *User) ExpireOTP(OTP string, db *gorm.DB) error {
	if !user.OtpExpiresAt.Before(time.Now()) && user.Otp == OTP {
		user.OtpExpiresAt = time.Now()
		return db.Save(user).Error
	}
	return fmt.Errorf("Invalid OTP")
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

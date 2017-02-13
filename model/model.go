package model

import (
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var DB *gorm.DB

func init() {
	db, err := gorm.Open("mysql", "conf:confdb@/confectionerydev?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db.DB().Ping()
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	db.AutoMigrate(&Content{})
	DB = db
}

type Model struct {
	ID        uint64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created"`
	UpdatedAt time.Time  `json:"updated"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

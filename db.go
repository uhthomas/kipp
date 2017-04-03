package conf

import "github.com/jinzhu/gorm"

type DB struct {
	*gorm.DB
}

func NewDB(dialect string, args ...interface{}) (*DB, error) {
	db, err := gorm.Open(dialect, args...)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(Content{}).Error; err != nil {
		return nil, err
	}
	return &DB{db}, nil
}

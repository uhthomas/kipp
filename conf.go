package conf

import (
	"time"

	"github.com/jinzhu/gorm"
)

// DB wraps gorm.DB which will allow for changing the way databases work in
// future.
type DB struct {
	*gorm.DB
}

// NewDB will create a new gorm DB and create tables associated with conf.
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

// Model is the base of all database models to ensure consistency.
type Model struct {
	ID        uint64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created"`
	UpdatedAt time.Time  `json:"updated"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

// Content is the model used for storing conf content information.
type Content struct {
	Model
	Name      string     `json:"name"`
	Extension string     `json:"extension"`
	Slug      string     `json:"slug" sql:"unique"`
	Hash      string     `json:"hash"`
	Size      int64      `json:"size"`
	Expires   *time.Time `json:"expires"`
	Key       []byte     `json:"-"`
	IV        []byte     `json:"-"`
}

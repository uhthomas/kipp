package conf

import (
	"time"

	"github.com/jinzhu/gorm"
)

//go:generate env GOARCH=js gopherjs build github.com/6f7262/conf/gopherjs -o default/public/js/gopher.js -m

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

type Model struct {
	ID        uint64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"created"`
	UpdatedAt time.Time  `json:"updated"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

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

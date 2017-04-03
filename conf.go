package conf

import "time"

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
	Slug      string     `json:"slug"`
	Hash      string     `json:"hash"`
	Size      int64      `json:"size"`
	Expires   *time.Time `json:"expires"`
	Key       []byte     `json:"-"`
	IV        []byte     `json:"-"`
}

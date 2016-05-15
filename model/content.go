package model

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/xid"
)

type Content struct {
	Model
	Name      string     `json:"name"`
	Extension string     `json:"extension"`
	Slug      string     `json:"slug"`
	Hash      string     `json:"hash"`
	Size      uint64     `json:"size"`
	Expires   *time.Time `json:"expires"`
	Key       []byte     `json:"-"`
	IV        []byte     `json:"-"`
}

func (c *Content) BeforeCreate(tx *gorm.DB) {
	c.Slug = xid.New().String()
}

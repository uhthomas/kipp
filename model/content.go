package model

import (
	"conf/crypto"
	"encoding/hex"
	"time"

	"github.com/jinzhu/gorm"
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
	// c.Slug = xid.New().String()
	b, _ := crypto.Random(10)
	c.Slug = hex.EncodeToString(b)
}

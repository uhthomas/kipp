package kipp

import "time"

type Content struct {
	ID        string
	Name      string
	Sum       string
	Size      uint64
	Lifetime  *time.Time
	Timestamp time.Time
}

type ContentStore interface {
	Create(c Content) error
	Fetch(id string) (Content, error)
}

package conf

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"
)

// Driver holds authentication information for a database driver.
type Driver struct {
	Dialect            string
	Username, Password string
	Path               string
}

// Open will open a new database connection given the driver information.
func (d Driver) Open() (*gorm.DB, error) {
	db, err := gorm.Open(d.Dialect, d.String())
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Content{}).Error; err != nil {
		return nil, err
	}
	return db, nil
}

// String constructs the appropriate authentication information for connection
// to a database.
func (d Driver) String() string {
	switch d.Dialect {
	case "postgres":
		return fmt.Sprintf("user='%s' pass='%s' host='%s' dbname=conf", d.Username, d.Password, d.Path)
	case "sqlite3":
		return d.Path
	case "mssql":
		return fmt.Sprintf("sqlserver://%s:%s@%s?database=conf", d.Username, d.Password, d.Path)
	case "mysql":
		return fmt.Sprintf("%s:%s@%s/conf", d.Username, d.Password, d.Path)
	}
	panic("conf: invalid driver")
}

// Content will store information about an uploaded file such as its name, hash,
// expiration date and slug.
type Content struct {
	Checksum  string
	CreatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
	Expires   *time.Time
	ID        []byte `gorm:"primary_key"`
	Name      string
	Size      int64
	UpdatedAt time.Time
}

// BeforeCreate will assign default values to the content. BeforeCreate is used
// by gorm.
func (c *Content) BeforeCreate() error {
	// make the length variable?
	c.ID = make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, c.ID)
	return err
}

package kipp

import (
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Driver holds authentication information for a database driver.
type Driver struct {
	Dialect            string
	Username, Password string
	Path               string
}

// Open will open a new database connection given the driver information.
func (d Driver) Open() (*sql.DB, error) {
	db, err := sql.Open(d.Dialect, d.String())
	if err != nil {
		return nil, err
	}
	// prevent "database is locked"
	if d.Dialect == "sqlite3" {
		db.SetMaxOpenConns(1)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS files (
		checksum char(86) NOT NULL,
		created_at timestamp DEFAULT CURRENT_TIMESTAMP,
		deleted_at timestamp,
		expires datetime,
		id char(15) NOT NULL,
		name varchar(255) NOT NULL,
		size bigint NOT NULL,
		PRIMARY KEY(id)
	)`); err != nil {
		return nil, err
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files(deleted_at)"); err != nil {
		return nil, err
	}
	return db, nil
}

// String constructs the appropriate authentication information for connection
// to a database.
func (d Driver) String() string {
	switch d.Dialect {
	case "postgres":
		return fmt.Sprintf("user='%s' pass='%s' host='%s' dbname=kipp", d.Username, d.Password, d.Path)
	case "sqlite3":
		return d.Path
	case "mssql":
		return fmt.Sprintf("sqlserver://%s:%s@%s?database=kipp", d.Username, d.Password, d.Path)
	case "mysql":
		return fmt.Sprintf("%s:%s@%s/kipp", d.Username, d.Password, d.Path)
	}
	panic("kipp: invalid driver")
}

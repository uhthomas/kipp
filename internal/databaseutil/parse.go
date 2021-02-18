package databaseutil

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/uhthomas/kipp/database"
	"github.com/uhthomas/kipp/database/badger"
	"github.com/uhthomas/kipp/database/sql"
)

// Parse parses s, and will create the appropriate database for the scheme.
func Parse(ctx context.Context, s string) (database.Database, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "":
		return badger.Open(u.Path)
	case "psql", "postgres", "postgresql":
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return sql.Open(ctx, "pgx", u.String())
	}
	return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
}

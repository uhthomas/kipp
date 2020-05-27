package databaseutil

import (
	"context"
	"fmt"
	"net/url"

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
	case "postgresql":
		u.Query().Set("sslmode", "require")
		return sql.Open(ctx, "postgresql", u.String())
	}
	return nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
}

package kipp

import (
	"context"
	"time"

	"github.com/uhthomas/kipp/database"
	"github.com/uhthomas/kipp/filesystem"
	"github.com/uhthomas/kipp/internal/databaseutil"
	"github.com/uhthomas/kipp/internal/filesystemutil"
)

type Option func(ctx context.Context, s *Server) error

func DB(db database.Database) Option {
	return func(ctx context.Context, s *Server) error {
		s.Database = db
		return nil
	}
}

func ParseDB(ss string) Option {
	return func(ctx context.Context, s *Server) error {
		db, err := databaseutil.Parse(ctx, ss)
		if err != nil {
			return err
		}
		return DB(db)(ctx, s)
	}
}

func FS(fs filesystem.FileSystem) Option {
	return func(ctx context.Context, s *Server) error {
		s.FileSystem = fs
		return nil
	}
}

func ParseFS(ss string) Option {
	return func(ctx context.Context, s *Server) error {
		fs, err := filesystemutil.Parse(ss)
		if err != nil {
			return err
		}
		return FS(fs)(ctx, s)
	}
}

func Lifetime(d time.Duration) Option {
	return func(ctx context.Context, s *Server) error {
		s.Lifetime = d
		return nil
	}
}

func Limit(n int64) Option {
	return func(ctx context.Context, s *Server) error {
		s.Limit = n
		return nil
	}
}

func Data(path string) Option {
	return func(ctx context.Context, s *Server) error {
		s.PublicPath = path
		return nil
	}
}

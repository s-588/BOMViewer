package db

import (
	"context"
	"database/sql"

	"github.com/s-588/BOMViewer/internal/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	queries *sqlite.Queries
	db      *sql.DB
}

func NewRepository(ctx context.Context,connStr string) (*Repository,error){
	conn, err := sql.Open("sqlite3",connStr)
	if err != nil {
		return nil,err
	}
	
	return &Repository{
		queries: sqlite.New(conn),
		db: conn,
	},err
}

func (r *Repository) Close() error {
	return r.db.Close()
}
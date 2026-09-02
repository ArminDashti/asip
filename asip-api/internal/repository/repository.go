package repository

import (
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type AsRepository struct {
	db *sql.DB
}

func NewAsRepository(db *sql.DB) *AsRepository {
	return &AsRepository{db: db}
}

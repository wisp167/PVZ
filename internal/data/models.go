package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	PVZ PVZModelv1
}

func NewModels(db *sql.DB) Models {
	return Models{
		PVZ: PVZModelv1{DB: db},
	}
}

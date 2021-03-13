package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type model struct {
	db        *pgxpool.Pool
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

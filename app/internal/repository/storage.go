package repository

import "github.com/jackc/pgx/v5/pgxpool"

type StoragePostgres struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *StoragePostgres {
	return &StoragePostgres{
		pool: pool,
	}
}

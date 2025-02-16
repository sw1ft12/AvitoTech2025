package repository

import (
	"AvitoTech/internal/models"
	"context"
	"github.com/jackc/pgx/v5"
	"log"
)

func (s *StoragePostgres) CreateUser(ctx context.Context, input models.AuthRequest) (models.User, error) {
	query := `INSERT INTO Users (username, password) VALUES (@username, @password) RETURNING username, password, coins`

	row, err := s.pool.Query(ctx, query, pgx.NamedArgs{
		"username": input.Username,
		"password": input.Password,
	})
	defer row.Close()

	if err != nil {
		return models.User{}, err
	}
	user, err := pgx.CollectOneRow(row, pgx.RowToStructByName[models.User])
	log.Println(err)
	return user, nil
}

package repository

import (
	"AvitoTech/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type InternalError struct {
	err string
}

func (e *InternalError) Error() string {
	return e.err
}

func (s *StoragePostgres) GetUserInfo(ctx context.Context, username string) (models.Info, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var coins int
	err = s.pool.QueryRow(ctx, `SELECT coins FROM Users WHERE username=@username`, pgx.NamedArgs{"username": username}).Scan(&coins)
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	rows, err := s.pool.Query(ctx, `SELECT "from", amount FROM History WHERE "to"=@username`, pgx.NamedArgs{"username": username})
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	receiveHistory, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ReceiveHistory])
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	rows, err = s.pool.Query(ctx, `SELECT "to", amount FROM History WHERE "from"=@username`, pgx.NamedArgs{"username": username})
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	sendHistory, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.SendHistory])
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	rows, err = s.pool.Query(ctx, `SELECT type, quantity FROM Inventory WHERE owner=@username`, pgx.NamedArgs{"username": username})
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	inventory, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ItemInfo])
	if err != nil {
		return models.Info{}, &InternalError{
			err: err.Error(),
		}
	}

	var info models.Info
	info.Inventory = inventory
	info.Coins = coins
	info.History.Received = receiveHistory
	info.History.Sent = sendHistory

	return info, nil
}

func (s *StoragePostgres) SendCoins(ctx context.Context, sender, receiver string, amount int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var senderBalance int
	err = tx.QueryRow(ctx, `SELECT coins FROM Users WHERE username=@username FOR UPDATE`, pgx.NamedArgs{"username": sender}).Scan(&senderBalance)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	var receiverBalance int
	err = tx.QueryRow(ctx, `SELECT coins FROM Users WHERE username=@username FOR UPDATE`, pgx.NamedArgs{"username": receiver}).Scan(&receiverBalance)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	if senderBalance < amount {
		return fmt.Errorf("not enough coins")
	}

	_, err = tx.Exec(ctx, `UPDATE Users SET coins = coins - @amount WHERE username = @username`, pgx.NamedArgs{"username": sender, "amount": amount})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	_, err = tx.Exec(ctx, `UPDATE Users SET coins = coins + @amount WHERE username=@username`, pgx.NamedArgs{"username": receiver, "amount": amount})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO History ("from", "to", amount) VALUES(@from, @to, @amount)`, pgx.NamedArgs{"from": sender, "to": receiver, "amount": amount})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	return nil
}

func (s *StoragePostgres) TopUpBalance(ctx context.Context, username string, amount int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `SELECT coins FROM Users WHERE username=@username FOR UPDATE`, pgx.NamedArgs{"username": username})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	_, err = tx.Exec(ctx, `UPDATE Users SET coins=coins+@amount WHERE username=@username`, pgx.NamedArgs{"username": username, "amount": amount})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}
	return nil
}

func (s *StoragePostgres) BuyItem(ctx context.Context, username, itemName string, itemCost int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	var balance int
	err = tx.QueryRow(ctx, `SELECT coins FROM Users WHERE username=@username FOR UPDATE`, pgx.NamedArgs{"username": username}).Scan(&balance)
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	if balance < itemCost {
		return fmt.Errorf("not enough coins")
	}

	_, err = tx.Exec(ctx, `UPDATE Users SET coins = coins - @cost WHERE username=@username`, pgx.NamedArgs{"username": username, "cost": itemCost})

	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	var quantity int
	err = tx.QueryRow(ctx, `SELECT quantity FROM Inventory WHERE owner=@owner AND type=@type FOR UPDATE`, pgx.NamedArgs{
		"type":     itemName,
		"quantity": 0,
		"owner":    username,
	}).Scan(&quantity)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = tx.Exec(ctx, `INSERT INTO Inventory (type, quantity, owner) VALUES(@type, @quantity, @owner)`, pgx.NamedArgs{
				"type":     itemName,
				"quantity": 1,
				"owner":    username,
			})
			if err != nil {
				return &InternalError{
					err: err.Error(),
				}
			}
			return nil
		} else {
			return &InternalError{
				err: err.Error(),
			}
		}
	}
	_, err = tx.Exec(ctx, `UPDATE Inventory SET quantity=quantity+1 WHERE owner=@owner AND type=@type`, pgx.NamedArgs{
		"owner": username,
		"type":  itemName,
	})
	if err != nil {
		return &InternalError{
			err: err.Error(),
		}
	}

	return nil
}

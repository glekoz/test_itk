package repository

import (
	"context"
	"errors"

	"github.com/glekoz/test_itk/internal/repository/db"
	"github.com/glekoz/test_itk/internal/shared/myerrors"
	"github.com/glekoz/test_itk/internal/shared/myvars"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
	p *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Repository{
		q: db.New(pool),
		p: pool,
	}, nil
}

func (r *Repository) CreateWallet(ctx context.Context, id string) error {
	err := r.q.CreateWallet(ctx, id)
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == UniqueViolationCode {
				return myerrors.ErrAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (r *Repository) GetBalance(ctx context.Context, id string) (int, error) {
	amount, err := r.q.GetBalance(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, myerrors.ErrNotFound
		}
		return 0, err
	}
	return int(amount), nil
}

func (r *Repository) Deposit(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) error {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	rowsAffected, err := qtx.Deposit(ctx, db.DepositParams{
		ID:     walletID,
		Amount: int32(amount),
	})
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return myerrors.ErrNotFound
	}

	err = qtx.CreateTransaction(ctx, db.CreateTransactionParams{
		ID:            transactionID,
		WalletID:      walletID,
		Amount:        int32(amount),
		OperationType: string(operationType),
	})
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == UniqueViolationCode {
				return myerrors.ErrAlreadyExists
			}
			if errp.Code == ForeignKeyViolationCode {
				return myerrors.ErrNotFound
			}
		}
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) Withdraw(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) error {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	rowsAffected, err := qtx.Withdraw(ctx, db.WithdrawParams{
		ID:     walletID,
		Amount: int32(amount),
	})
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == CheckViolationCode {
				return myerrors.ErrNegativeAmount
			}
		}
		return err
	}
	if rowsAffected != 1 {
		return myerrors.ErrNotFound
	}

	err = qtx.CreateTransaction(ctx, db.CreateTransactionParams{
		ID:            transactionID,
		WalletID:      walletID,
		Amount:        int32(amount),
		OperationType: string(operationType),
	})
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == UniqueViolationCode {
				return myerrors.ErrAlreadyExists
			}
			if errp.Code == ForeignKeyViolationCode {
				return myerrors.ErrNotFound
			}
		}
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) Close() {
	r.p.Close()
}

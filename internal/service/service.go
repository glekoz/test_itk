package service

import (
	"context"
	"log"

	"github.com/glekoz/test_itk/internal/shared/myvars"
	"github.com/google/uuid"
)

type RepoAPI interface {
	CreateWallet(ctx context.Context, id string) error
	GetBalance(ctx context.Context, id string) (int, error)
	Deposit(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error)
	Withdraw(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error)
}

type CacheAPI interface {
	Add(walletID string, balance int) error
	Get(walletID string) (int, bool)
	Delete(walletID string)
}

type Service struct {
	repo     RepoAPI
	cache    CacheAPI
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New(repo RepoAPI, cache CacheAPI, infoLog, errorLog *log.Logger) *Service {
	return &Service{
		repo:     repo,
		cache:    cache,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (a *Service) CreateWallet(ctx context.Context) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		a.errorLog.Printf("new uuid creating failed")
		return "", err
	}
	idstr := id.String()
	err = a.repo.CreateWallet(ctx, idstr)
	if err != nil {
		return "", err
	}
	return idstr, nil
}

func (a *Service) GetBalance(ctx context.Context, walletID string) (int, error) {
	balance, ok := a.cache.Get(walletID)
	if ok {
		return balance, nil
	}
	balance, err := a.repo.GetBalance(ctx, walletID)
	if err != nil {
		return 0, err
	}
	if err := a.cache.Add(walletID, balance); err != nil {
		a.errorLog.Printf("adding to cache failed")
	}
	return balance, nil
}

func (a *Service) Deposit(ctx context.Context, walletID string, amount int) error {
	transactionID, err := uuid.NewV7()
	if err != nil {
		a.errorLog.Printf("new uuid creating failed")
		return err
	}
	balance, err := a.repo.Deposit(ctx, walletID, transactionID.String(), amount, myvars.OperationTypeDeposit)
	if err != nil {
		return err
	}
	if err := a.cache.Add(walletID, balance); err != nil {
		a.errorLog.Printf("adding to cache failed")
		a.cache.Delete(walletID)
	}

	return nil
}

func (a *Service) Withdraw(ctx context.Context, walletID string, amount int) error {
	transactionID, err := uuid.NewV7()
	if err != nil {
		a.errorLog.Printf("new uuid creating failed")
		return err
	}

	balance, err := a.repo.Withdraw(ctx, walletID, transactionID.String(), amount, myvars.OperationTypeWithdraw)
	if err != nil {
		return err
	}

	if err := a.cache.Add(walletID, balance); err != nil {
		a.cache.Delete(walletID)
	}

	return nil
}

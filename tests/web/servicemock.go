package web

import (
	"context"
	"errors"
)

type MockService struct {
	CreateWalletFunc func(ctx context.Context) (string, error)
	GetBalanceFunc   func(ctx context.Context, walletID string) (int, error)
	DepositFunc      func(ctx context.Context, walletID string, amount int) error
	WithdrawFunc     func(ctx context.Context, walletID string, amount int) error
}

func (m *MockService) CreateWallet(ctx context.Context) (string, error) {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc(ctx)
	}
	return "", errors.New("not implemented")
}

func (m *MockService) GetBalance(ctx context.Context, walletID string) (int, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, walletID)
	}
	return 0, errors.New("not implemented")
}

func (m *MockService) Deposit(ctx context.Context, walletID string, amount int) error {
	if m.DepositFunc != nil {
		return m.DepositFunc(ctx, walletID, amount)
	}
	return errors.New("not implemented")
}

func (m *MockService) Withdraw(ctx context.Context, walletID string, amount int) error {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, walletID, amount)
	}
	return errors.New("not implemented")
}

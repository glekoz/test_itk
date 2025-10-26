package service_test

import (
	"context"
	"log"
	"os"

	"github.com/glekoz/test_itk/internal/shared/myvars"
)

// MockRepo представляет мок для репозитория
type MockRepo struct {
	CreateWalletFunc func(ctx context.Context, id string) error
	GetBalanceFunc   func(ctx context.Context, id string) (int, error)
	DepositFunc      func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error)
	WithdrawFunc     func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error)
}

func (m *MockRepo) CreateWallet(ctx context.Context, id string) error {
	if m.CreateWalletFunc != nil {
		return m.CreateWalletFunc(ctx, id)
	}
	return nil
}

func (m *MockRepo) GetBalance(ctx context.Context, id string) (int, error) {
	if m.GetBalanceFunc != nil {
		return m.GetBalanceFunc(ctx, id)
	}
	return 0, nil
}

func (m *MockRepo) Deposit(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
	if m.DepositFunc != nil {
		return m.DepositFunc(ctx, walletID, transactionID, amount, operationType)
	}
	return 0, nil
}

func (m *MockRepo) Withdraw(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
	if m.WithdrawFunc != nil {
		return m.WithdrawFunc(ctx, walletID, transactionID, amount, operationType)
	}
	return 0, nil
}

// MockCache представляет мок для кэша
type MockCache struct {
	AddFunc    func(walletID string, balance int) error
	GetFunc    func(walletID string) (int, bool)
	DeleteFunc func(walletID string)
}

func (m *MockCache) Add(walletID string, balance int) error {
	if m.AddFunc != nil {
		return m.AddFunc(walletID, balance)
	}
	return nil
}

func (m *MockCache) Get(walletID string) (int, bool) {
	if m.GetFunc != nil {
		return m.GetFunc(walletID)
	}
	return 0, false
}

func (m *MockCache) Delete(walletID string) {
	if m.DeleteFunc != nil {
		m.DeleteFunc(walletID)
	}
}

// testHelpers содержит вспомогательные функции для тестов
type testHelpers struct {
	infoLog  *log.Logger
	errorLog *log.Logger
}

func newTestHelpers() *testHelpers {
	return &testHelpers{
		infoLog:  log.New(os.Stdout, "TEST INFO: ", log.LstdFlags),
		errorLog: log.New(os.Stdout, "TEST ERROR: ", log.LstdFlags),
	}
}

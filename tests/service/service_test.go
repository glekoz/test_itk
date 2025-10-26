package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/glekoz/test_itk/internal/service"
	"github.com/glekoz/test_itk/internal/shared/myvars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateWallet(t *testing.T) {
	helpers := newTestHelpers()

	tests := []struct {
		name          string
		repoMock      *MockRepo
		cacheMock     *MockCache
		expectedError bool
		errorContains string
	}{
		{
			name: "successful wallet creation",
			repoMock: &MockRepo{
				CreateWalletFunc: func(ctx context.Context, id string) error {
					// Проверяем что ID передается корректно
					assert.NotEmpty(t, id)
					return nil
				},
			},
			cacheMock:     &MockCache{},
			expectedError: false,
		},
		{
			name: "repository error on creation",
			repoMock: &MockRepo{
				CreateWalletFunc: func(ctx context.Context, id string) error {
					return errors.New("database connection failed")
				},
			},
			cacheMock:     &MockCache{},
			expectedError: true,
			errorContains: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service.New(tt.repoMock, tt.cacheMock, helpers.infoLog, helpers.errorLog)

			walletID, err := service.CreateWallet(context.Background())

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, walletID)
				// Проверяем что ID является валидным UUID
				assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`, walletID)
			}
		})
	}
}

func TestService_GetBalance(t *testing.T) {
	helpers := newTestHelpers()

	tests := []struct {
		name            string
		walletID        string
		repoMock        *MockRepo
		cacheMock       *MockCache
		expectedBalance int
		expectedError   bool
		errorContains   string
	}{
		{
			name:     "successful balance from cache",
			walletID: "cached-wallet",
			cacheMock: &MockCache{
				GetFunc: func(walletID string) (int, bool) {
					assert.Equal(t, "cached-wallet", walletID)
					return 1500, true
				},
			},
			repoMock:        &MockRepo{},
			expectedBalance: 1500,
			expectedError:   false,
		},
		{
			name:     "successful balance from repository",
			walletID: "db-wallet",
			cacheMock: &MockCache{
				GetFunc: func(walletID string) (int, bool) {
					return 0, false // кэш пустой
				},
			},
			repoMock: &MockRepo{
				GetBalanceFunc: func(ctx context.Context, id string) (int, error) {
					assert.Equal(t, "db-wallet", id)
					return 2000, nil
				},
			},
			expectedBalance: 2000,
			expectedError:   false,
		},
		{
			name:     "repository error on balance retrieval",
			walletID: "error-wallet",
			cacheMock: &MockCache{
				GetFunc: func(walletID string) (int, bool) {
					return 0, false
				},
			},
			repoMock: &MockRepo{
				GetBalanceFunc: func(ctx context.Context, id string) (int, error) {
					return 0, errors.New("wallet not found")
				},
			},
			expectedBalance: 0,
			expectedError:   true,
			errorContains:   "wallet not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := service.New(tt.repoMock, tt.cacheMock, helpers.infoLog, helpers.errorLog)

			balance, err := service.GetBalance(context.Background(), tt.walletID)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBalance, balance)
			}
		})
	}
}

func TestService_Deposit(t *testing.T) {
	helpers := newTestHelpers()

	tests := []struct {
		name          string
		walletID      string
		amount        int
		repoMock      *MockRepo
		cacheMock     *MockCache
		expectedError bool
		errorContains string
		cacheCalled   bool
	}{
		{
			name:     "successful deposit",
			walletID: "test-wallet",
			amount:   1000,
			repoMock: &MockRepo{
				DepositFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					assert.Equal(t, "test-wallet", walletID)
					assert.Equal(t, 1000, amount)
					assert.Equal(t, myvars.OperationTypeDeposit, operationType)
					assert.NotEmpty(t, transactionID)
					return 1500, nil // новый баланс
				},
			},
			cacheMock: &MockCache{
				AddFunc: func(walletID string, balance int) error {
					assert.Equal(t, "test-wallet", walletID)
					assert.Equal(t, 1500, balance)
					return nil
				},
			},
			expectedError: false,
			cacheCalled:   true,
		},
		{
			name:     "repository error on deposit",
			walletID: "error-wallet",
			amount:   1000,
			repoMock: &MockRepo{
				DepositFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					return 0, errors.New("insufficient funds")
				},
			},
			cacheMock:     &MockCache{},
			expectedError: true,
			errorContains: "insufficient funds",
			cacheCalled:   false,
		},
		{
			name:     "cache update error after successful deposit",
			walletID: "test-wallet",
			amount:   1000,
			repoMock: &MockRepo{
				DepositFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					return 1500, nil
				},
			},
			cacheMock: &MockCache{
				AddFunc: func(walletID string, balance int) error {
					return errors.New("cache update failed")
				},
				DeleteFunc: func(walletID string) {
					assert.Equal(t, "test-wallet", walletID)
					// Должен быть вызван при ошибке добавления в кэш
				},
			},
			expectedError: false, // Ошибка кэша не должна влиять на успешность операции
			cacheCalled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheAddCalled := false

			// Сохраняем оригинальные функции и отслеживаем вызовы
			if tt.cacheMock.AddFunc != nil {
				originalAdd := tt.cacheMock.AddFunc
				tt.cacheMock.AddFunc = func(walletID string, balance int) error {
					cacheAddCalled = true
					return originalAdd(walletID, balance)
				}
			}

			if tt.cacheMock.DeleteFunc != nil {
				originalDelete := tt.cacheMock.DeleteFunc
				tt.cacheMock.DeleteFunc = func(walletID string) {
					originalDelete(walletID)
				}
			}

			service := service.New(tt.repoMock, tt.cacheMock, helpers.infoLog, helpers.errorLog)

			err := service.Deposit(context.Background(), tt.walletID, tt.amount)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.cacheCalled {
				assert.True(t, cacheAddCalled, "Cache Add should have been called")
			}
		})
	}
}

func TestService_Withdraw(t *testing.T) {
	helpers := newTestHelpers()

	tests := []struct {
		name          string
		walletID      string
		amount        int
		repoMock      *MockRepo
		cacheMock     *MockCache
		expectedError bool
		errorContains string
		cacheCalled   bool
	}{
		{
			name:     "successful withdraw",
			walletID: "test-wallet",
			amount:   500,
			repoMock: &MockRepo{
				WithdrawFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					assert.Equal(t, "test-wallet", walletID)
					assert.Equal(t, 500, amount)
					assert.Equal(t, myvars.OperationTypeWithdraw, operationType)
					assert.NotEmpty(t, transactionID)
					return 500, nil // новый баланс
				},
			},
			cacheMock: &MockCache{
				AddFunc: func(walletID string, balance int) error {
					assert.Equal(t, "test-wallet", walletID)
					assert.Equal(t, 500, balance)
					return nil
				},
			},
			expectedError: false,
			cacheCalled:   true,
		},
		{
			name:     "repository error on withdraw",
			walletID: "error-wallet",
			amount:   1000,
			repoMock: &MockRepo{
				WithdrawFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					return 0, errors.New("insufficient funds")
				},
			},
			cacheMock:     &MockCache{},
			expectedError: true,
			errorContains: "insufficient funds",
			cacheCalled:   false,
		},
		{
			name:     "cache update error after successful withdraw",
			walletID: "test-wallet",
			amount:   500,
			repoMock: &MockRepo{
				WithdrawFunc: func(ctx context.Context, walletID, transactionID string, amount int, operationType myvars.OperationType) (int, error) {
					return 500, nil
				},
			},
			cacheMock: &MockCache{
				AddFunc: func(walletID string, balance int) error {
					return errors.New("cache update failed")
				},
				DeleteFunc: func(walletID string) {
					assert.Equal(t, "test-wallet", walletID)
				},
			},
			expectedError: false, // Ошибка кэша не должна влиять на успешность операции
			cacheCalled:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheAddCalled := false

			if tt.cacheMock.AddFunc != nil {
				originalAdd := tt.cacheMock.AddFunc
				tt.cacheMock.AddFunc = func(walletID string, balance int) error {
					cacheAddCalled = true
					return originalAdd(walletID, balance)
				}
			}

			service := service.New(tt.repoMock, tt.cacheMock, helpers.infoLog, helpers.errorLog)

			err := service.Withdraw(context.Background(), tt.walletID, tt.amount)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tt.cacheCalled {
				assert.True(t, cacheAddCalled, "Cache Add should have been called")
			}
		})
	}
}

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/glekoz/test_itk/api/v1"
	"github.com/glekoz/test_itk/internal/shared/myerrors"
	"github.com/glekoz/test_itk/internal/web/v1"
)

func TestServer_CreateWallet(t *testing.T) {
	tests := []struct {
		name             string
		mockFunc         func(ctx context.Context) (string, error)
		expectedStatus   int
		expectedLocation string
	}{
		{
			name: "successful creation",
			mockFunc: func(ctx context.Context) (string, error) {
				return "test-wallet-123", nil
			},
			expectedStatus:   http.StatusCreated,
			expectedLocation: "http://test-host/api/v1/wallets/test-wallet-123",
		},
		{
			name: "service error",
			mockFunc: func(ctx context.Context) (string, error) {
				return "", errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{
				CreateWalletFunc: tt.mockFunc,
			}

			// Используем nil логгеры для тестов, или можно создать буферизованные логгеры
			server := web.New(mockService, "test-host", log.Default(), log.Default())

			req := httptest.NewRequest("POST", "/api/v1/wallets", nil)
			w := httptest.NewRecorder()

			server.CreateWallet(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedLocation != "" {
				location := resp.Header.Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected location %s, got %s", tt.expectedLocation, location)
				}
			}
		})
	}
}

func TestServer_GetBalance(t *testing.T) {
	tests := []struct {
		name           string
		walletID       string
		mockFunc       func(ctx context.Context, walletID string) (int, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful balance retrieval",
			walletID: "existing-wallet",
			mockFunc: func(ctx context.Context, walletID string) (int, error) {
				return 1000, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"balance":1000,"wallet_id":"existing-wallet"}`,
		},
		{
			name:     "wallet not found",
			walletID: "non-existing-wallet",
			mockFunc: func(ctx context.Context, walletID string) (int, error) {
				return 0, myerrors.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "empty wallet id",
			walletID:       "",
			mockFunc:       nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "internal server error",
			walletID: "error-wallet",
			mockFunc: func(ctx context.Context, walletID string) (int, error) {
				return 0, errors.New("internal error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{
				GetBalanceFunc: tt.mockFunc,
			}

			server := web.New(mockService, "test-host", log.Default(), log.Default())

			url := fmt.Sprintf("/api/v1/wallets/%s/balance", tt.walletID)
			req := httptest.NewRequest("GET", url, nil)

			// Устанавливаем path parameter для gorilla/mux или стандартного роутера
			if tt.walletID != "" {
				req.SetPathValue("wallet_uuid", tt.walletID)
			}

			w := httptest.NewRecorder()

			server.GetBalance(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				var balanceResp api.Balance
				if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}

				expectedBalance := api.Balance{Balance: 1000, WalletId: "existing-wallet"}
				if balanceResp != expectedBalance {
					t.Errorf("expected body %+v, got %+v", expectedBalance, balanceResp)
				}
			}
		})
	}
}

func TestServer_Transfer(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockDeposit    func(ctx context.Context, walletID string, amount int) error
		mockWithdraw   func(ctx context.Context, walletID string, amount int) error
		expectedStatus int
	}{
		{
			name: "successful deposit",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    1000,
				Operation: api.Deposit,
			},
			mockDeposit: func(ctx context.Context, walletID string, amount int) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "successful withdraw",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    500,
				Operation: api.Withdraw,
			},
			mockWithdraw: func(ctx context.Context, walletID string, amount int) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "deposit - wallet not found",
			requestBody: api.Transfer{
				WalletId:  "non-existing",
				Amount:    1000,
				Operation: api.Deposit,
			},
			mockDeposit: func(ctx context.Context, walletID string, amount int) error {
				return myerrors.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "withdraw - insufficient funds",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    5000,
				Operation: api.Withdraw,
			},
			mockWithdraw: func(ctx context.Context, walletID string, amount int) error {
				return myerrors.ErrNegativeAmount
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "withdraw - wallet not found",
			requestBody: api.Transfer{
				WalletId:  "non-existing",
				Amount:    500,
				Operation: api.Withdraw,
			},
			mockWithdraw: func(ctx context.Context, walletID string, amount int) error {
				return myerrors.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing wallet_id",
			requestBody: api.Transfer{
				Amount:    1000,
				Operation: api.Deposit,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid amount",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    -100,
				Operation: api.Deposit,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid operation",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    1000,
				Operation: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal server error on deposit",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    1000,
				Operation: api.Deposit,
			},
			mockDeposit: func(ctx context.Context, walletID string, amount int) error {
				return errors.New("internal error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "internal server error on withdraw",
			requestBody: api.Transfer{
				WalletId:  "test-wallet",
				Amount:    500,
				Operation: api.Withdraw,
			},
			mockWithdraw: func(ctx context.Context, walletID string, amount int) error {
				return errors.New("internal error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockService{
				DepositFunc:  tt.mockDeposit,
				WithdrawFunc: tt.mockWithdraw,
			}

			server := web.New(mockService, "test-host", log.Default(), log.Default())

			var bodyBytes []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/v1/transfer", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			server.Transfer(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

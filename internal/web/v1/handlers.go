package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/glekoz/test_itk/api/v1"
	"github.com/glekoz/test_itk/internal/shared/myerrors"
)

type ServiceAPI interface {
	CreateWallet(ctx context.Context) (string, error)
	GetBalance(ctx context.Context, walletID string) (int, error)
	Deposit(ctx context.Context, walletID string, amount int) error
	Withdraw(ctx context.Context, walletID string, amount int) error
}

type Server struct {
	service  ServiceAPI
	host     string
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New(service ServiceAPI, h string, infoLog, errorLog *log.Logger) *Server {
	return &Server{
		service:  service,
		host:     h,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (s *Server) Transfer(w http.ResponseWriter, r *http.Request) {
	var req api.Transfer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.infoLog.Printf("%s - %s %s %s - ended with error (400, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), "invalid JSON body")
		SendError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	var errs string
	if req.WalletId == "" {
		errs += "wallet_id is required; "
	}
	if req.Amount <= 0 {
		errs += "amount must be positive; "
	}
	if req.Operation != api.Deposit && req.Operation != api.Withdraw {
		errs += "operation must be either 'deposit' or 'withdraw'."
	}
	if len(errs) > 0 {
		s.infoLog.Printf("%s - %s %s %s - ended with error (400, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), errs)
		SendError(w, http.StatusBadRequest, errs)
		return
	}

	switch req.Operation {
	case api.Deposit:
		err := s.service.Deposit(r.Context(), req.WalletId, req.Amount)
		if err != nil {
			if errors.Is(err, myerrors.ErrNotFound) {
				s.infoLog.Printf("%s - %s %s %s - ended with error (404, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
				SendError(w, http.StatusNotFound, err.Error())
				return
			}
			s.infoLog.Printf("%s - %s %s %s - ended with error (500, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	case api.Withdraw:
		err := s.service.Withdraw(r.Context(), req.WalletId, req.Amount)
		if err != nil {
			if errors.Is(err, myerrors.ErrNegativeAmount) {
				s.infoLog.Printf("%s - %s %s %s - ended with error (400, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
				SendError(w, http.StatusBadRequest, err.Error())
				return
			} else if errors.Is(err, myerrors.ErrNotFound) {
				s.infoLog.Printf("%s - %s %s %s - ended with error (404, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
				SendError(w, http.StatusNotFound, err.Error())
				return
			}
			s.infoLog.Printf("%s - %s %s %s - ended with error (500, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.infoLog.Printf("%s - %s %s %s - completed successful", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) CreateWallet(w http.ResponseWriter, r *http.Request) {
	res, err := s.service.CreateWallet(r.Context())
	if err != nil {
		s.infoLog.Printf("%s - %s %s %s - ended with error (500, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
		SendError(w, http.StatusInternalServerError, err.Error()) // репозиторий может вернуть AlreadyExists, но в данном случае это считаем ошибкой сервера
		return
	}
	s.infoLog.Printf("%s - %s %s %s - completed successful", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
	w.Header().Set("Location", fmt.Sprintf("http://%s/api/v1/wallets/%s", s.host, res))
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) GetBalance(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("wallet_uuid")
	if walletID == "" {
		s.infoLog.Printf("%s - %s %s %s - ended with error (400, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), "wallet uuid can not be empty")
		SendError(w, http.StatusBadRequest, "wallet uuid can not be empty")
		return
	}

	res, err := s.service.GetBalance(r.Context(), walletID)
	if err != nil {
		if errors.Is(err, myerrors.ErrNotFound) {
			s.infoLog.Printf("%s - %s %s %s - ended with error (404, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
			SendError(w, http.StatusNotFound, err.Error())
			return
		}
		s.infoLog.Printf("%s - %s %s %s - ended with error (500, %s)", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI(), err.Error())
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.infoLog.Printf("%s - %s %s %s - completed successful", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
	WriteJSON(w, http.StatusOK, api.Balance{
		Balance:  res,
		WalletId: walletID,
	})
}

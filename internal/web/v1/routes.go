package web

import (
	"net/http"

	"github.com/justinas/alice"
)

func (a *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/wallet", a.Transfer)
	mux.HandleFunc("POST /api/v1/wallet/create", a.CreateWallet)
	mux.HandleFunc("GET /api/v1/wallets/{wallet_uuid}", a.GetBalance)

	standard := alice.New(a.recoverPanic, a.logRequest)
	return standard.Then(mux)
}

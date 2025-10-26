package web

import (
	"encoding/json"
	"net/http"

	"github.com/glekoz/test_itk/api/v1"
)

func SendError(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	erro := json.NewEncoder(w).Encode(api.Error{
		Status: status,
		Title:  http.StatusText(status),
		Detail: err,
	})
	if erro != nil {
		http.Error(w, "Failed to send error response", http.StatusInternalServerError)
	}
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

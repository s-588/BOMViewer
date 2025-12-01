package helpers

import (
	"log/slog"
	"net/http"
)

func WriteAndLogError(w http.ResponseWriter, statusCode int, message string, err error) {
	slog.Error(message, "error", err)
	w.WriteHeader(statusCode)
	w.Write([]byte("Ошибка: " + err.Error()))
}

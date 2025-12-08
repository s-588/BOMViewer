package helpers

import (
	"encoding/base64"
	"log/slog"
	"net/http"
)

func SetAndLogError(w http.ResponseWriter, statusCode int, msg string, log string, opts ...any) {
	slog.Error(log, opts...)
	w.Header().Add("HX-Error", toBase64(msg))
	w.WriteHeader(statusCode)
}

func SetAndLogSuccess(w http.ResponseWriter, msg string, log string, opts ...any) {
	slog.Info(log, opts...)
	w.Header().Add("HX-Alert", toBase64(msg))
	w.WriteHeader(http.StatusOK)
}

func SetAndLogAlert(w http.ResponseWriter,status int, msg string, log string, opts ...any) {
	slog.Warn(log, opts...)
	w.Header().Add("HX-Alert", toBase64(msg))
	w.WriteHeader(status)
}

func toBase64(msg string) string{
	return base64.StdEncoding.EncodeToString([]byte(msg))
}
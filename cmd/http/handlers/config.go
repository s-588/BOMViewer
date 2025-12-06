package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/s-588/BOMViewer/cmd/config"
	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/web/templates"
)

func (h *Handler) ConfigPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.SettingsForm(h.cfg).Render(r.Context(), w)
}

func (h *Handler) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	port, err := strconv.Atoi(r.FormValue("server.server_port"))
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "incorrect server port", errors.New("Неверный формат порта сервера"))
		return
	}
	if port < 1 || port > 65535 {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "server port out of range", errors.New("Порт сервера должен быть в диапазоне от 1 до 65535"))
		return
	}
	if !slices.Contains([]string{"DEBUG", "INFO", "WARN", "ERROR"}, r.FormValue("log.log_level")) {
		helpers.WriteAndLogError(w, http.StatusBadRequest, "invalid log level", errors.New("Недопустимый уровень логирования"))
		return
	}
	cfg := config.Config{
		BaseDirectory: filepath.Clean(r.FormValue("base_directory")),
		ServerCfg: config.ServerConfig{
			ServerPort: int(port),
			UploadsDir: filepath.Clean(r.FormValue("server.uploads_directory")),
		},
		DBCfg: config.DBConfig{
			DBName: filepath.Clean(r.FormValue("database.database_name")),
		},
		LogCfg: config.LogConfig{
			LogLevel: r.FormValue("log.log_level"),
		},
	}
	err = h.cfg.UpdateConfig(cfg)
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "error updating config", err)
		return
	}
	slog.Info("configuration updated", "newConfig", cfg)
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (h *Handler) ResetConfigHandler(w http.ResponseWriter, r *http.Request) {
	err := h.cfg.ResetConfig()
	if err != nil {
		helpers.WriteAndLogError(w, http.StatusInternalServerError, "error resetting config", err)
		return
	}
	slog.Info("configuration reseted")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

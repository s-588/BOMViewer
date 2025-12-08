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
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) ConfigPageHandler(w http.ResponseWriter, r *http.Request) {
	templates.SettingsForm(h.cfg).Render(r.Context(), w)
}

func (h *Handler) UpdateConfigHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	port, err := strconv.Atoi(r.FormValue("server.server_port"))
	if err != nil {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Неверный формат порта сервера", "invalid server port format in update config handler", err)
		return
	}
	if (port < 1024 && port != 0) || port > 49151 {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Порт сервера вне диапазона", "server port out of range in update config handler", errors.New("Порт сервера должен быть в диапазоне от 1024 до 49151 или равен 0"))
		return
	}
	if !slices.Contains([]string{"DEBUG", "INFO", "WARN", "ERROR"}, r.FormValue("log.log_level")) {
		helpers.SetAndLogError(w, http.StatusBadRequest, "Недопустимый уровень логирования", "invalid log level in update config handler", errors.New("Недопустимый уровень логирования"))
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
	if r.FormValue("web_ui_password") != "" {
		if r.FormValue("web_ui_password") != r.FormValue("web_ui_password_confirm") {
			helpers.SetAndLogError(w, http.StatusBadRequest, "Пароли не совпадают", "passwords do not match in update config handler", errors.New("Пароли не совпадают"))
			return
		}
		passHash, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("web_ui_password")), bcrypt.DefaultCost)
		if err != nil {
			helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка хеширования пароля", "error hashing password in update config handler", err)
			return
		}
		cfg.WebUIPassword = string(passHash)
	}
	slog.Debug("config before updating", "cfg", h.cfg)
	err = h.cfg.UpdateConfig(cfg)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка обновления конфигурации", "error updating config in update config handler", err)
		return
	}
	slog.Info("configuration updated", "newConfig", cfg)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Конфигурация сохранена"))
}

func (h *Handler) ResetConfigHandler(w http.ResponseWriter, r *http.Request) {
	if field := r.PathValue("field"); field != "" {
		err := h.cfg.ResetField(field)
		if err != nil {
			helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка сброса поля конфигурации", "error resetting config field in reset config handler", err)
			return
		}
		templates.SettingsField(field, h.cfg).Render(r.Context(), w)
		return
	}
	err := h.cfg.ResetConfig()
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка сброса конфигурации", "error resetting config in reset config handler", err)
		return
	}
	slog.Info("configuration reseted")
	templates.SettingsForm(h.cfg).Render(r.Context(),w)
}

package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/s-588/BOMViewer/internal/helpers"
	"github.com/s-588/BOMViewer/web/templates"
	"golang.org/x/crypto/bcrypt"
)

type AuthManager struct {
	passHash string
	sessions map[string]time.Time
	mu       sync.RWMutex
}

func NewAuthManager(passHash string) *AuthManager {
	return &AuthManager{
		passHash: passHash,
		sessions: make(map[string]time.Time),
	}
}

func (m *AuthManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.passHash == "" {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/login" ||
			strings.HasPrefix(r.URL.Path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		session, err := r.Cookie("session")
		if err != nil {
			templates.LoginPage().Render(r.Context(), w)
			return
		}

		t, ok := m.sessions[session.Value]
		if !ok || time.Now().After(t) {
			delete(m.sessions, session.Value)
			templates.LoginPage().Render(r.Context(), w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *AuthManager) LoginHandler(w http.ResponseWriter, r *http.Request) {
	pass := r.FormValue("password")
	if pass == "" {
		helpers.SetAndLogError(w, http.StatusUnauthorized, "пустой пароль", "empty password attempt", "error", errors.New("Неверный пароль"))
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(m.passHash), []byte(pass)); err != nil {
		helpers.SetAndLogError(w, http.StatusUnauthorized, "неверный пароль", "invalid password attempt", "error", err)
		return
	}
	sessionIDData := make([]byte, 32)
	_, err := rand.Read(sessionIDData)
	if err != nil {
		helpers.SetAndLogError(w, http.StatusInternalServerError, "ошибка генерации идентификатора сессии", "error generating session ID", "error", err)
		return
	}
	sessionID := base64.URLEncoding.EncodeToString(sessionIDData)
	m.mu.Lock()
	m.sessions[string(sessionID)] = time.Now().Add(24 * time.Hour)
	m.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		HttpOnly: true,
	})

	slog.Info("user successfully logged in", "session_id", sessionID)
	w.Header().Add("HX-Redirect", "/welcome")
	w.WriteHeader(http.StatusOK)
}

func (m *AuthManager) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	m.mu.Lock()
	delete(m.sessions, session.Value)
	m.mu.Unlock()
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

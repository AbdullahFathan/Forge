package auth

import (
	"net"
	"net/http"
	"time"

	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/response"
	"workspace/pkg/validator"
)

const RefreshCookie = "refresh_token"

type Handler struct {
	svc          *Service
	cookieSecure bool
	refreshTTL   time.Duration
}

func NewHandler(svc *Service, cookieSecure bool, refreshTTL time.Duration) *Handler {
	return &Handler{svc: svc, cookieSecure: cookieSecure, refreshTTL: refreshTTL}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	AccessToken string      `json:"accessToken"`
	ExpiresIn   int         `json:"expiresIn"`
	User        user.Public `json:"user"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	tok, err := h.svc.Login(r.Context(), req.Email, req.Password, clientIP(r))
	if err != nil {
		response.Error(w, err)
		return
	}
	h.setRefreshCookie(w, tok.Refresh)
	response.JSON(w, http.StatusOK, loginResponse{
		AccessToken: tok.AccessToken,
		ExpiresIn:   tok.ExpiresIn,
		User:        user.ToPublic(tok.User),
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(RefreshCookie)
	if err != nil || c.Value == "" {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	tok, err := h.svc.Refresh(r.Context(), c.Value)
	if err != nil {
		h.clearRefreshCookie(w)
		response.Error(w, err)
		return
	}
	h.setRefreshCookie(w, tok.Refresh)
	response.JSON(w, http.StatusOK, loginResponse{
		AccessToken: tok.AccessToken,
		ExpiresIn:   tok.ExpiresIn,
		User:        user.ToPublic(tok.User),
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(RefreshCookie)
	token := ""
	if c != nil {
		token = c.Value
	}
	_ = h.svc.Logout(r.Context(), token)
	h.clearRefreshCookie(w)
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) setRefreshCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    value,
		Path:     "/auth",
		MaxAge:   int(h.refreshTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

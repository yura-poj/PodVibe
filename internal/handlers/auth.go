package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"podvibe/internal/services"
)

type AuthHandler struct {
	auth *services.AuthService
}

func NewAuthHandler(auth *services.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if !BindJSON(c, &req) {
		return
	}
	res, err := h.auth.Register(req.Email, req.Username, req.Password, req.DisplayName)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"user":          sanitizeUser(res.User),
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !BindJSON(c, &req) {
		return
	}
	res, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		status := http.StatusUnauthorized
		if err == services.ErrUserBanned {
			status = http.StatusForbidden
		}
		RespondError(c, status, "unauthorized", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":          sanitizeUser(res.User),
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if !BindJSON(c, &req) {
		return
	}
	res, err := h.auth.Refresh(req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		if err == services.ErrUserBanned {
			status = http.StatusForbidden
		}
		RespondError(c, status, "unauthorized", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  res.AccessToken,
		"refresh_token": res.RefreshToken,
	})
}

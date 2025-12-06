package services

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"podvibe/internal/auth"
	"podvibe/internal/config"
	"podvibe/internal/models"
	"podvibe/internal/repositories"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBanned         = errors.New("user is banned")
)

type AuthService struct {
	users  *repositories.UserRepository
	tokens *repositories.TokenRepository
	cfg    config.Config
}

func NewAuthService(users *repositories.UserRepository, tokens *repositories.TokenRepository, cfg config.Config) *AuthService {
	return &AuthService{users: users, tokens: tokens, cfg: cfg}
}

type AuthResult struct {
	User         *models.User
	AccessToken  string
	RefreshToken string
}

func (s *AuthService) Register(email, username, password, displayName string) (*AuthResult, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        email,
		Username:     s.users.NormalizeUsername(username),
		PasswordHash: string(hashed),
		DisplayName:  displayName,
		Role:         "user",
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return s.generateTokens(user)
}

func (s *AuthService) Login(email, password string) (*AuthResult, error) {
	user, err := s.users.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.IsBanned {
		return nil, ErrUserBanned
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.generateTokens(user)
}

func (s *AuthService) Refresh(refreshToken string) (*AuthResult, error) {
	valid, stored, err := s.tokens.IsValid(refreshToken)
	if err != nil {
		return nil, err
	}
	if !valid || stored == nil {
		return nil, ErrInvalidCredentials
	}
	claims, err := auth.ParseToken(refreshToken, s.cfg.JWTSecret)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	user, err := s.users.FindByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	if user.IsBanned {
		return nil, ErrUserBanned
	}
	// revoke old token to allow logout behaviour
	if err := s.tokens.Revoke(refreshToken); err != nil {
		return nil, err
	}
	return s.generateTokens(user)
}

func (s *AuthService) generateTokens(user *models.User) (*AuthResult, error) {
	access, refresh, err := auth.GenerateTokens(user.ID, user.Role, s.cfg.JWTSecret, s.cfg.AccessTokenTTL, s.cfg.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Save(user.ID, refresh, time.Now().Add(s.cfg.RefreshTokenTTL)); err != nil {
		return nil, err
	}
	return &AuthResult{
		User:         user,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

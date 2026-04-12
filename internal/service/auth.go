package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/amigoer/kite/internal/config"
	"github.com/amigoer/kite/internal/model"
	"github.com/amigoer/kite/internal/repo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUserExists         = errors.New("username or email already exists")
	ErrRegistrationClosed = errors.New("user registration is not allowed")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("invalid token")
)

// JWTClaims JWT 令牌中携带的声明。
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair 包含 access token 和 refresh token。
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// AuthService 认证业务逻辑。
type AuthService struct {
	userRepo     *repo.UserRepo
	tokenRepo    *repo.APITokenRepo
	cfg          config.AuthConfig
}

func NewAuthService(userRepo *repo.UserRepo, tokenRepo *repo.APITokenRepo, cfg config.AuthConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		cfg:       cfg,
	}
}

// Register 用户自行注册。
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	if !s.cfg.AllowRegistration {
		return nil, ErrRegistrationClosed
	}

	exists, err := s.userRepo.ExistsByUsernameOrEmail(ctx, username, email)
	if err != nil {
		return nil, fmt.Errorf("register check exists: %w", err)
	}
	if exists {
		return nil, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("register hash password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("register create user: %w", err)
	}

	return user, nil
}

// Login 用户登录，返回 JWT token pair。
func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// 也尝试通过邮箱登录
		user, err = s.userRepo.GetByEmail(ctx, username)
		if err != nil {
			return nil, ErrInvalidCredentials
		}
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokenPair(user)
}

// RefreshToken 使用 refresh token 获取新的 token pair。
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}

	return s.generateTokenPair(user)
}

// ValidateToken 验证 JWT token 并返回 claims。
func (s *AuthService) ValidateToken(tokenStr string) (*JWTClaims, error) {
	return s.parseToken(tokenStr)
}

// ValidateAPIToken 验证 API Token，返回关联的用户 ID。
func (s *AuthService) ValidateAPIToken(ctx context.Context, tokenStr string) (string, error) {
	hash := HashToken(tokenStr)

	token, err := s.tokenRepo.GetByTokenHash(ctx, hash)
	if err != nil {
		return "", ErrTokenInvalid
	}

	if token.IsExpired() {
		return "", ErrTokenExpired
	}

	// 异步更新最后使用时间
	go func() {
		_ = s.tokenRepo.UpdateLastUsed(context.Background(), token.ID)
	}()

	return token.UserID, nil
}

// CreateAPIToken 为用户创建 API Token。
// 返回明文 token（仅此一次展示）和记录。
func (s *AuthService) CreateAPIToken(ctx context.Context, userID, name string, expiresAt *time.Time) (string, *model.APIToken, error) {
	plainToken, err := generateRandomToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("generate api token: %w", err)
	}

	token := &model.APIToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		TokenHash: HashToken(plainToken),
		ExpiresAt: expiresAt,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return "", nil, fmt.Errorf("create api token: %w", err)
	}

	return plainToken, token, nil
}

// CreateAdminUser 创建管理员用户（安装向导使用）。
func (s *AuthService) CreateAdminUser(ctx context.Context, username, email, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash admin password: %w", err)
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "admin",
		StorageLimit: -1, // 管理员无存储限制
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create admin user: %w", err)
	}

	return user, nil
}

func (s *AuthService) generateTokenPair(user *model.User) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(s.cfg.AccessTokenExpiry)

	accessClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	refreshClaims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
	}, nil
}

func (s *AuthService) parseToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// HashToken 计算 token 的 SHA256 哈希。
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateRandomToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"youmind-backend-v2/internal/config"
	"youmind-backend-v2/internal/domain/user"
)

// Service 认证应用服务
type Service struct {
	cfg  *config.Config
	repo user.Repository
}

func NewService(cfg *config.Config, repo user.Repository) *Service {
	return &Service{cfg: cfg, repo: repo}
}

// RegisterInput 注册入参
type RegisterInput struct {
	Username string
	Email    string
	Password string
}

// RegisterResult 注册结果
type RegisterResult struct {
	ID               string
	Username         string
	Email            string
	SubscriptionPlan string
	CreditsBalance   int
	CreditsUsed      int
	CreatedAt        time.Time
}

// Register 注册新用户
func (s *Service) Register(in RegisterInput) (*RegisterResult, error) {
	exists, err := s.repo.ExistsByUsernameOrEmail(in.Username, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameOrEmailTaken
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &user.User{
		ID:               uuid.NewString(),
		Username:         in.Username,
		Email:            in.Email,
		HashedPassword:   string(hashed),
		SubscriptionPlan: "free",
		CreditsBalance:   1000,
		CreditsUsed:      0,
		CreatedAt:        now,
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return &RegisterResult{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		SubscriptionPlan: u.SubscriptionPlan,
		CreditsBalance:   u.CreditsBalance,
		CreditsUsed:      u.CreditsUsed,
		CreatedAt:        u.CreatedAt,
	}, nil
}

// LoginInput 登录入参
type LoginInput struct {
	Username string
	Password string
}

// LoginResult 登录结果（仅含 token，用户信息由 /me 拉取）
type LoginResult struct {
	AccessToken string
	TokenType   string
}

// Login 验证账号密码并返回 JWT
func (s *Service) Login(in LoginInput) (*LoginResult, error) {
	u, err := s.repo.GetByUsername(in.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(in.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	expires := time.Now().Add(time.Duration(s.cfg.AccessTokenExpireMin) * time.Minute)
	token, err := s.createJWT(u.ID, expires)
	if err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: token, TokenType: "bearer"}, nil
}

// GetUserByID 供中间件/me 使用
func (s *Service) GetUserByID(id string) (*user.User, error) {
	return s.repo.GetByID(id)
}

// UpdateProfile 更新用户名/邮箱
func (s *Service) UpdateProfile(id string, username, email *string) (*user.User, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if username != nil {
		u.Username = *username
	}
	if email != nil {
		u.Email = *email
	}
	if err := s.repo.Update(u); err != nil {
		return nil, err
	}
	return u, nil
}

// Secret 返回 JWT 密钥，供 HTTP 中间件校验 token 使用
func (s *Service) Secret() string {
	return s.cfg.JWTSecret
}

func (s *Service) createJWT(sub string, expires time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": sub,
		"exp": expires.Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

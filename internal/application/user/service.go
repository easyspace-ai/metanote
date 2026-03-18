package user

import (
	"youmind-backend-v2/internal/domain/user"
)

// Service 用户应用服务（余额等）
type Service struct {
	repo user.Repository
}

func NewService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

// GetBalance 获取当前用户余额信息
func (s *Service) GetBalance(userID string) (*BalanceResult, error) {
	u, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return &BalanceResult{
		Plan:    u.SubscriptionPlan,
		Balance: u.CreditsBalance,
		Used:    u.CreditsUsed,
	}, nil
}

// BalanceResult 余额返回
type BalanceResult struct {
	Plan    string `json:"plan"`
	Balance int    `json:"balance"`
	Used    int    `json:"used"`
}

package persistence

import (
	"youmind-backend-v2/internal/domain/user"
)

// UserRepository 用户仓储 GORM 实现
type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) user.Repository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *user.User) error {
	m := toUserModel(u)
	return r.db.Create(m).Error
}

func (r *UserRepository) GetByID(id string) (*user.User, error) {
	var m UserModel
	if err := r.db.Where("id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepository) GetByUsername(username string) (*user.User, error) {
	var m UserModel
	if err := r.db.Where("username = ?", username).First(&m).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepository) GetByEmail(email string) (*user.User, error) {
	var m UserModel
	if err := r.db.Where("email = ?", email).First(&m).Error; err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *UserRepository) ExistsByUsernameOrEmail(username, email string) (bool, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("username = ? OR email = ?", username, email).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) Update(u *user.User) error {
	m := toUserModel(u)
	return r.db.Save(m).Error
}

func toUserModel(u *user.User) *UserModel {
	return &UserModel{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		HashedPassword:   u.HashedPassword,
		SubscriptionPlan: u.SubscriptionPlan,
		CreditsBalance:   u.CreditsBalance,
		CreditsUsed:      u.CreditsUsed,
		CreatedAt:        u.CreatedAt,
	}
}

func toUserEntity(m *UserModel) *user.User {
	return &user.User{
		ID:               m.ID,
		Username:         m.Username,
		Email:            m.Email,
		HashedPassword:   m.HashedPassword,
		SubscriptionPlan: m.SubscriptionPlan,
		CreditsBalance:   m.CreditsBalance,
		CreditsUsed:      m.CreditsUsed,
		CreatedAt:        m.CreatedAt,
	}
}

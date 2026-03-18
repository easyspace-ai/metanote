package user

import "time"

// User 用户聚合根
type User struct {
	ID               string
	Username         string
	Email            string
	HashedPassword   string
	SubscriptionPlan string
	CreditsBalance   int
	CreditsUsed      int
	CreatedAt        time.Time
}

package user

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (u *UserApp) GetUserByPublicID(id string) (*domain.User, error) {
	user, err := u.users.GetByPublicID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		logger.LogError(err.Error(), "user-app")
		return nil, domain.Failed("failed to get user")
	}

	if user == nil {
		return nil, domain.NotFound("user not found")
	}

	if user.MlKemPublicKey == "" || user.EcdhPublicKey == "" || user.EdPublicKey == "" {
		return nil, domain.NotFound("user not found")
	}

	return user, nil
}

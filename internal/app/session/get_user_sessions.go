package session

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (s *SessionApp) GetUserSessions(user_id int) ([]*domain.Session, error) {
	user, err := s.users.GetByID(user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		logger.LogError(err.Error(), "session-app")
		return nil, domain.Failed("failed to get user")
	}
	if user == nil {
		return nil, domain.NotFound("user not found")
	}

	sessions, err := s.session.GetByUserID(user.ID)
	if err != nil {
		logger.LogError(err.Error(), "session-app")
		return nil, domain.Failed("failed to get user sessions")
	}

	if len(sessions) == 0 {
		return nil, domain.NotFound("sessions not found")
	}

	return sessions, nil
}

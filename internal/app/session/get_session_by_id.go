package session

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (s *SessionApp) GetSessionByID(id int) (*domain.Session, error) {
	session, err := s.session.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("session not found")
		}
		logger.LogError(err.Error(), "session-app")
		return nil, domain.Failed("failed to get session")
	}

	if session == nil {
		return nil, domain.NotFound("session not found")
	}

	return session, nil
}

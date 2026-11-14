package chat

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (c *ChatApp) GetMemberGroups(ctx context.Context, memberID int) ([]domain.GroupMember, error) {
	groups, err := c.groups.GetByMemberID(ctx, memberID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.GroupMember{}, nil
		}
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to get groups")
	}

	if groups == nil {
		groups = []domain.GroupMember{}
	}

	return groups, nil
}

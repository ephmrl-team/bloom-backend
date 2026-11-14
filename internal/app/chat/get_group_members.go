package chat

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (c *ChatApp) GetGroupMembers(ctx context.Context, groupID int) ([]domain.GroupMember, error) {
	members, err := c.groups.GetByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.GroupMember{}, nil
		}
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to get group members")
	}

	if members == nil {
		members = []domain.GroupMember{}
	}

	return members, nil
}

package chat

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
)

func (c *ChatApp) GetGroupMember(ctx context.Context, memberID, groupID int) (*domain.GroupMember, error) {
	member, err := c.groups.GetByMemberAndChatID(ctx, memberID, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("member not found")
		}
		return nil, domain.Failed("failed to get group members")
	}

	return member, nil
}

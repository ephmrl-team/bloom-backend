package chat

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
)

func (c *ChatApp) HasMember(ctx context.Context, chat *domain.Chat, memberID int) (bool, error) {
	switch chat.Type {
	case "private":
		for _, m := range chat.Members {
			if m.ID == memberID {
				return true, nil
			}
		}
	case "group":
		_, err := c.groups.GetByMemberAndChatID(ctx, memberID, chat.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	default:
		return false, nil
	}
	return false, nil
}

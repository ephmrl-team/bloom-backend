package chat

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (c *ChatApp) GetChatByID(ctx context.Context, user_id int, id int) (*domain.Chat, error) {
	chat, err := c.chats.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("chat not found")
		}
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to get chat")
	}

	if chat == nil {
		return nil, domain.NotFound("chat not found")
	}

	isMember, err := c.HasMember(ctx, chat, user_id)
	if err != nil {
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to check chat membership")
	}
	if !isMember {
		return nil, domain.NotFound("chat not found")
	}

	return chat, nil
}

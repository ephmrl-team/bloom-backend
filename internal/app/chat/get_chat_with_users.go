package chat

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (c *ChatApp) GetChatWithUsers(user_id int, recipient int) (*domain.Chat, error) {
	chat, err := c.chats.GetWithUsers(user_id, recipient)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to get chat")
	}

	return chat, nil
}

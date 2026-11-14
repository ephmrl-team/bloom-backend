package message

import (
	"context"
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (m *MessageApp) GetChatLastReadMessage(ctx context.Context, user_id, chatID int) (*domain.Message, error) {
	chat, err := m.chats.GetChatByID(ctx, user_id, chatID)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			return nil, appErr
		}
		logger.LogError(err.Error(), "chat-app")
		return nil, domain.Failed("failed to get chat")
	}

	message, err := m.messages.GetChatLastReadMessage(chat.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("last read message not found")
		}
		logger.LogError(err.Error(), "message-app")
		return nil, domain.Failed("failed to get last read message")
	}

	if message == nil {
		return nil, domain.NotFound("last read message not found")
	}

	return message, nil
}

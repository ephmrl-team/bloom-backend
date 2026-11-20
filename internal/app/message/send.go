package message

import (
	"context"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
	"github.com/slipe-fun/skid-backend/internal/pointer"
)

func (m *MessageApp) Send(ctx context.Context, user_id int, message *domain.SocketMessage) (*domain.MessageWithReply, *domain.Chat, bool, error) {
	chat, err := m.chats.GetChatByID(ctx, user_id, message.ChatID)
	if err != nil {
		return nil, nil, false, err
	}

	var replyTo *domain.Message
	if message.ReplyTo != 0 {
		reply_to_message, err := m.messages.GetByID(message.ReplyTo)
		if err != nil || reply_to_message == nil || reply_to_message.ChatID != chat.ID {
			return nil, nil, false, domain.NotFound("reply to message not found")
		}
		replyTo = reply_to_message
	}

	newMessage, replayed, err := m.messages.Create(&domain.Message{
		Version:         message.Version,
		Ciphertext:      message.Ciphertext,
		Nonce:           message.Nonce,
		Salt:            message.Salt,
		ChatID:          message.ChatID,
		ReplyTo:         pointer.Intptr(message.ReplyTo),
		ClientMessageID: pointer.Strptr(message.ClientMessageID),
	})
	if err != nil {
		logger.LogError(err.Error(), "message-app")
		return nil, nil, false, domain.Failed("failed to create message")
	}

	createdMessage := &domain.MessageWithReply{
		Message:        *newMessage,
		ReplyToMessage: replyTo,
	}

	return createdMessage, chat, replayed, nil
}

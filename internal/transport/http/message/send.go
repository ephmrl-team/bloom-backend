package message

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto/validations"
)

func (h *MessageHandler) Send(c *fiber.Ctx) error {
	userVal := c.Locals("session_user")
	sessionUser, ok := userVal.(*domain.User)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req struct {
		Version    string `json:"version"`
		Ciphertext string `json:"ciphertext"`
		Nonce      string `json:"nonce"`
		Salt       string `json:"salt"`
		ChatID     int    `json:"chat_id"`
		ReplyTo    int    `json:"reply_to"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request",
		})
	}

	if req.Ciphertext == "" || req.Nonce == "" || req.Salt == "" || req.ChatID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request",
		})
	}

	if err := crypto.ValidateVersion(req.Version); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "unsupported_protocol_version",
			"message": "unsupported protocol version",
		})
	}

	if err := validations.ValidateCryptoLength(req.Nonce, crypto.NonceSize); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_nonce_length",
			"message": "invalid nonce length",
		})
	}

	if err := validations.ValidateCryptoLength(req.Salt, crypto.SaltSize); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_salt_length",
			"message": "invalid salt length",
		})
	}

	message, chat, err := h.messageApp.Send(c.Context(), sessionUser.ID, &domain.SocketMessage{
		Version:    req.Version,
		Ciphertext: req.Ciphertext,
		Nonce:      req.Nonce,
		Salt:       req.Salt,
		ChatID:     req.ChatID,
		ReplyTo:    req.ReplyTo,
	})
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Msg,
		})
	}

	outMsg := struct {
		Type    string          `json:"type"`
		ID      int             `json:"id"`
		UserID  string          `json:"user_id"`
		ReplyTo *domain.Message `json:"reply_to,omitempty"`
		*domain.MessageWithReply
	}{
		Type:             "message.new",
		ID:               message.ID,
		UserID:           sessionUser.PublicID,
		ReplyTo:          message.ReplyToMessage,
		MessageWithReply: message,
	}

	b, err := json.Marshal(outMsg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "internal error",
		})
	}

	switch chat.Type {
	case "private":
		if other := h.chatApp.GetOtherMember(chat, sessionUser.ID); other != nil {
			h.wsHub.SendToUser(other.ID, b)
		}
	case "group":
		members, err := h.chatApp.GetGroupMembers(c.Context(), chat.ID)
		if appErr, ok := err.(*domain.AppError); ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"error":   appErr.Code,
				"message": appErr.Msg,
			})
		}
		for i := range members {
			member := members[i]
			h.wsHub.SendToUser(member.MemberID, b)
		}
	}

	return c.JSON(message)
}

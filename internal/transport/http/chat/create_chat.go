package chat

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto/validations"
)

func (h *ChatHandler) CreateChat(c *fiber.Ctx) error {
	userVal := c.Locals("session_user")
	sessionUser, ok := userVal.(*domain.User)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var req struct {
		Type string `json:"type"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request",
		})
	}

	switch req.Type {
	case "private":
		return h.createPrivateChat(c, sessionUser)
	case "group":
		return h.createGroupChat(c, sessionUser)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid chat type",
		})
	}
}

func (h *ChatHandler) createGroupChat(c *fiber.Ctx, sessionUser *domain.User) error {
	var req struct {
		Title   string   `json:"title"`
		Members []Member `json:"members"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request body",
		})
	}

	if len(req.Title) < 1 || len(req.Title) > 30 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid title length",
		})
	}

	if len(req.Members) > 100 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "too many members",
		})
	}

	groupMembers, buildErr := h.BuildGroupMemberList(MembersRequest{
		Members: req.Members,
	}, sessionUser)
	if buildErr != nil {
		return c.Status(buildErr.Status).JSON(fiber.Map{
			"error":   buildErr.Code,
			"message": buildErr.Msg,
		})
	}

	creatorGroupMember := domain.GroupMember{
		MemberID: sessionUser.ID,
		Role:     "creator",
	}
	groupMembers = append(groupMembers, creatorGroupMember)

	chat, err := h.chatApp.CreateGroupChat(c.Context(), req.Title, groupMembers, sessionUser.ID)
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Msg,
		})
	}

	for i := range groupMembers {
		member := groupMembers[i]

		outMsg := struct {
			*domain.Chat
			Type                     string      `json:"type"`
			ChatType                 string      `json:"chat_type"`
			UserID                   string      `json:"user_id"`
			Role                     string      `json:"role"`
			InvitedBy                domain.User `json:"invited_by"`
			domain.Handshake         `json:"handshake"`
			domain.EncryptedGroupKey `json:"encrypted_group_key"`
		}{
			Chat:              chat,
			Type:              "chat.new",
			ChatType:          "group",
			UserID:            sessionUser.PublicID,
			Role:              "member",
			InvitedBy:         *sessionUser,
			Handshake:         member.Handshake,
			EncryptedGroupKey: member.EncryptedGroupKey,
		}

		b, err := json.Marshal(outMsg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "internal_error",
				"message": "internal error",
			})
		}

		h.wsHub.SendToUser(member.MemberID, b)
	}

	return c.JSON(fiber.Map{
		"id":         chat.ID,
		"type":       chat.Type,
		"title":      chat.Title,
		"role":       "creator",
		"created_at": chat.CreatedAt,
	})
}

func (h *ChatHandler) createPrivateChat(c *fiber.Ctx, sessionUser *domain.User) error {
	var req struct {
		Recipient string `json:"recipient"`
		Handshake struct {
			Version             string `json:"version"`
			ReceiverCipherText  string `json:"receiver_ciphertext"`
			SenderEphemeralX448 string `json:"sender_ephemeral_x448"`
			SenderSignature     string `json:"sender_signature"`
			EncryptedSyncKey    struct {
				CipherText string `json:"ciphertext"`
				Nonce      string `json:"nonce"`
			} `json:"encrypted_sync_key"`
		} `json:"handshake"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request body",
		})
	}

	if req.Recipient == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_recipient",
			"message": "no recipient",
		})
	}

	if req.Handshake.Version == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_version",
			"message": "protocol version is missing",
		})
	}

	if err := crypto.ValidateVersion(req.Handshake.Version); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "unsupported_protocol_version",
			"message": "unsupported protocol version",
		})
	}

	if req.Handshake.SenderSignature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_sender_signature",
			"message": "sender signature is missing",
		})
	}

	if req.Handshake.ReceiverCipherText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_receiver_cipher_text",
			"message": "receiver cipher text is missing",
		})
	}

	if req.Handshake.SenderEphemeralX448 == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_sender_cipher_text",
			"message": "sender cipher text is missing",
		})
	}

	if req.Handshake.EncryptedSyncKey.CipherText == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_sync_key_ciphertext",
			"message": "encrypted sync key ciphertext is missing",
		})
	}

	if req.Handshake.EncryptedSyncKey.Nonce == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_sync_key_nonce",
			"message": "encrypted sync key nonce is missing",
		})
	}

	if err := validations.ValidateCryptoLength(req.Handshake.ReceiverCipherText, crypto.MLKEM768CiphertextSize); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_receiver_ciphertext_length"})
	}

	if err := validations.ValidateCryptoLength(req.Handshake.SenderEphemeralX448, crypto.X448PublicKeySize); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_sender_ciphertext_length"})
	}

	if err := validations.ValidateCryptoLength(req.Handshake.SenderSignature, crypto.SignatureSize); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_sender_signature_length"})
	}

	if err := validations.ValidateCryptoLength(req.Handshake.EncryptedSyncKey.Nonce, crypto.NonceSize); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_nonce_length"})
	}

	if err := validations.ValidateCryptoLength(req.Handshake.EncryptedSyncKey.CipherText, crypto.SyncKeyCiphertextSize); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid_sync_key_length"})
	}

	user, err := h.userApp.GetUserByPublicID(req.Recipient)
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Msg,
		})
	}

	if user.ID == sessionUser.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "self_chat_not_allowed",
		})
	}

	chat, err := h.chatApp.GetChatWithUsers(sessionUser.ID, user.ID)
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Msg,
		})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "internal error",
		})
	}

	if chat != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":   "already_exists",
			"message": "chat with users already exists",
		})
	}

	chat, err = h.chatApp.CreatePrivateChat(sessionUser.ID, user.ID, domain.Handshake{
		Version:             req.Handshake.Version,
		ReceiverCipherText:  req.Handshake.ReceiverCipherText,
		SenderEphemeralX448: req.Handshake.SenderEphemeralX448,
		SenderSignature:     req.Handshake.SenderSignature,
		EncryptedSyncKey: domain.EncryptedSyncKey{
			CipherText: req.Handshake.EncryptedSyncKey.CipherText,
			Nonce:      req.Handshake.EncryptedSyncKey.Nonce,
		},
	})
	if appErr, ok := err.(*domain.AppError); ok {
		return c.Status(appErr.Status).JSON(fiber.Map{
			"error":   appErr.Code,
			"message": appErr.Msg,
		})
	}

	outMsg := struct {
		*domain.Chat
		Type     string `json:"type"`
		ChatType string `json:"chat_type"`
		UserID   string `json:"user_id"`
	}{
		Chat:     chat,
		Type:     "chat.new",
		ChatType: "private",
		UserID:   sessionUser.PublicID,
	}

	b, err := json.Marshal(outMsg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "internal error",
		})
	}

	h.wsHub.SendToUser(sessionUser.ID, b)
	h.wsHub.SendToUser(user.ID, b)

	return c.JSON(chat)
}

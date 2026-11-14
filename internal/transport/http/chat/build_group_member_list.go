package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto/validations"
)

type Member struct {
	MemberID          string                   `json:"member_id"`
	Handshake         domain.Handshake         `json:"handshake"`
	EncryptedGroupKey domain.EncryptedGroupKey `json:"encrypted_group_key"`
}

type MembersRequest struct {
	Members []Member `json:"members"`
}

func (h *ChatHandler) BuildGroupMemberList(req MembersRequest, sessionUser *domain.User) ([]domain.GroupMember, *domain.AppError) {
	var members []domain.User
	users := make(map[string]domain.User, len(req.Members))
	var err error
	if len(req.Members) != 0 {
		var membersIDs []string
		for i := range req.Members {
			member := req.Members[i]
			membersIDs = append(membersIDs, member.MemberID)
		}

		members, err = h.userApp.GetUsersByPublicIDs(membersIDs)
		if err != nil {
			if appErr, ok := err.(*domain.AppError); ok {
				if appErr.Status == fiber.StatusNotFound {
					return nil, domain.InvalidRequest("invalid members list")
				}

				return nil, appErr
			}

			return nil, domain.InternalError("internal error")
		}

		if len(members) == 0 {
			return nil, domain.InvalidRequest("invalid members list")
		}

		for i := range members {
			member := members[i]
			users[member.PublicID] = member
		}
	}

	var groupMembers []domain.GroupMember

	for i := range req.Members {
		memberObject := req.Members[i]

		member, exists := users[memberObject.MemberID]
		if !exists {
			return nil, domain.UserNotFound("member " + memberObject.MemberID + " not found")
		}

		if member.ID == sessionUser.ID {
			continue
		}

		if memberObject.Handshake.ReceiverCipherText == "" {
			return nil, domain.NoReceiverCipherText(memberObject.MemberID + ": receiver cipher text is missing")
		}

		if memberObject.Handshake.SenderEphemeralX448 == "" {
			return nil, domain.NoSenderCipherText(memberObject.MemberID + ": sender cipher text is missing")
		}

		if memberObject.Handshake.EncryptedSyncKey.CipherText == "" {
			return nil, domain.NoSyncKeyCiphertext(memberObject.MemberID + ": encrypted sync key ciphertext is missing")
		}

		if memberObject.Handshake.EncryptedSyncKey.Nonce == "" {
			return nil, domain.NoSyncKeyNonce(memberObject.MemberID + ": encrypted sync key nonce is missing")
		}

		if err := validations.ValidateCryptoLength(memberObject.Handshake.ReceiverCipherText, MLKEM768CiphertextSize); err != nil {
			return nil, domain.InvalidReceiverCiphertextLength(memberObject.MemberID + ": invalid receiver ciphertext length")
		}

		if err := validations.ValidateCryptoLength(memberObject.Handshake.SenderEphemeralX448, X448PublicKeySize); err != nil {
			return nil, domain.InvalidSenderCiphertextLength(memberObject.MemberID + ": invalid sender ciphertext length")
		}

		if err := validations.ValidateCryptoLength(memberObject.Handshake.EncryptedSyncKey.Nonce, AESGCMNonceSize); err != nil {
			return nil, domain.InvalidNonceLength(memberObject.MemberID + ": invalid nonce length")
		}

		if err := validations.ValidateCryptoLength(memberObject.Handshake.EncryptedSyncKey.CipherText, SyncKeyCiphertextSize); err != nil {
			return nil, domain.InvalidSyncKeyLength(memberObject.MemberID + ": invalid sync key length")
		}

		newGroupMember := domain.GroupMember{
			MemberID:          member.ID,
			Role:              "member",
			Handshake:         memberObject.Handshake,
			EncryptedGroupKey: memberObject.EncryptedGroupKey,
		}

		groupMembers = append(groupMembers, newGroupMember)
	}

	return groupMembers, nil
}

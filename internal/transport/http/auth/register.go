package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
)

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.KeysRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid request",
		})
	}

	if req.AuthLookupID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "missing auth lookup ID",
		})
	}

	pubKeys := req.IdentityKeys.IdentityPublicKeys
	if len(pubKeys.MlKem768) == 0 || len(pubKeys.X448) == 0 || len(pubKeys.Ed448) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "missing required public keys",
		})
	}

	idSecret := req.IdentityKeys.EncryptedSecretKeys
	if len(idSecret.Version) == 0 || len(idSecret.Ciphertext) == 0 || len(idSecret.Nonce) == 0 || len(idSecret.Salt) == 0 || len(idSecret.Signature) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "missing required encrypted identity key fields",
		})
	}

	masterKey := req.EncryptedMasterKey
	if len(masterKey.Version) == 0 || len(masterKey.Ciphertext) == 0 || len(masterKey.Nonce) == 0 || len(masterKey.Salt) == 0 || len(masterKey.Signature) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "missing required encrypted master key fields",
		})
	}

	if err := crypto.ValidateVersionBytes(idSecret.Version); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "unsupported_protocol_version",
			"message": "unsupported identity key protocol version",
		})
	}

	if err := crypto.ValidateVersionBytes(masterKey.Version); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "unsupported_protocol_version",
			"message": "unsupported master key protocol version",
		})
	}

	if len(req.AuthLookupID) < 16 || len(req.AuthLookupID) > 22 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "invalid auth lookup ID",
		})
	}

	token, user, session, err := h.authApp.Register(&req)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			return c.Status(appErr.Status).JSON(fiber.Map{
				"error":   appErr.Code,
				"message": appErr.Msg,
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "internal_error",
			"message": "failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":                   user.PublicID,
			"username":             user.Username,
			"display_name":         user.DisplayName,
			"description":          user.Description,
			"ml_kem768_public_key": user.MlKemPublicKey,
			"x448_public_key":      user.EcdhPublicKey,
			"ed448_public_key":     user.EdPublicKey,
			"date":                 user.Date,
		},
		"session": fiber.Map{
			"id":         session.ID,
			"token":      session.Token,
			"user_id":    user.PublicID,
			"revoked_at": session.RevokedAt,
			"created_at": session.CreatedAt,
		},
	})
}

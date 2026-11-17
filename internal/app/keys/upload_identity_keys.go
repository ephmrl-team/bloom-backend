package keys

import (
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (k *KeysApp) UploadIdentityKeys(user_id int, req *domain.IdentityKeysRequest) error {
	user, err := k.users.GetByID(user_id)
	if err != nil || user == nil {
		return domain.Failed("user not found")
	}

	if user.MlKemPublicKey != "" || user.EcdhPublicKey != "" || user.EdPublicKey != "" {
		return domain.InvalidData("user already have keys")
	}

	if err := crypto.ValidateVersionBytes(req.EncryptedSecretKeys.Version); err != nil {
		return domain.InvalidData("unsupported protocol version")
	}

	publicID := identity.GenerateUserID(
		req.IdentityPublicKeys.Ed448,
		req.IdentityPublicKeys.X448,
		req.IdentityPublicKeys.MlKem768,
	)

	if publicID != user.PublicID {
		return domain.InvalidData("public keys do not match user ID")
	}

	if err := (&identity.User{ID: publicID, PublicKeys: req.IdentityPublicKeys}).Validate(); err != nil {
		return domain.InvalidData("invalid public keys")
	}

	if len(req.EncryptedSecretKeys.Signature) != crypto.SignatureSize {
		return domain.InvalidData("invalid signature length")
	}

	if len(req.EncryptedSecretKeys.Nonce) != crypto.NonceSize {
		return domain.InvalidData("invalid nonce")
	}

	if len(req.EncryptedSecretKeys.Salt) != crypto.SaltSize {
		return domain.InvalidData("invalid salt")
	}

	if len(req.EncryptedSecretKeys.Ciphertext) > 10000 {
		return domain.InvalidData("ciphertext too large")
	}

	isValid, err := crypto.VerifyIdentityKeysSignature(req.IdentityPublicKeys, &req.EncryptedSecretKeys)
	if err != nil {
		return domain.Failed("error during verification")
	}
	if !isValid {
		return domain.InvalidData("invalid identity signature")
	}

	_, err = k.keys.Create(&domain.EncryptedKeys{
		UserID: user_id,
		Type:   "bundle",
		EncryptedKey: domain.EncryptedKey{
			Version:    crypto.EncodeBase64(req.EncryptedSecretKeys.Version),
			Ciphertext: crypto.EncodeBase64(req.EncryptedSecretKeys.Ciphertext),
			Nonce:      crypto.EncodeBase64(req.EncryptedSecretKeys.Nonce),
			Salt:       crypto.EncodeBase64(req.EncryptedSecretKeys.Salt),
			Signature:  crypto.EncodeBase64(req.EncryptedSecretKeys.Signature),
		},
	})
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		return domain.Failed("failed to save encrypted bundle")
	}

	err = k.users.UpdatePublicKeys(
		user_id,
		crypto.EncodeBase64(req.IdentityPublicKeys.MlKem768),
		crypto.EncodeBase64(req.IdentityPublicKeys.X448),
		crypto.EncodeBase64(req.IdentityPublicKeys.Ed448),
	)
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		_ = k.keys.Delete(user_id)
		return domain.Failed("failed to update user public keys")
	}

	return nil
}

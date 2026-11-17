package keys

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (k *KeysApp) UploadMasterKey(user_id int, key *identity.EncryptedMasterKey) error {
	user, err := k.users.GetByID(user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NotFound("user not found")
		}
		logger.LogError(err.Error(), "keys-app")
		return domain.Failed("failed to get user")
	}
	if user == nil {
		return domain.NotFound("user not found")
	}

	_, err = k.keys.GetByUserID(user_id, "master")
	if err == nil {
		return domain.AlreadyExists("key is already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		logger.LogError(err.Error(), "keys-app")
		return domain.Failed("failed to get key")
	}

	if err := crypto.ValidateVersionBytes(key.Version); err != nil {
		return domain.InvalidData("unsupported protocol version")
	}

	if len(key.Ciphertext) > 10000 {
		return domain.InvalidData("ciphertext too large")
	}

	if len(key.Nonce) != crypto.NonceSize {
		return domain.InvalidData("invalid nonce")
	}

	if len(key.Salt) != crypto.SaltSize {
		return domain.InvalidData("invalid salt")
	}

	if len(key.Signature) != crypto.SignatureSize {
		return domain.InvalidData("invalid signature length")
	}

	publicKeys, err := crypto.DecodePublicKeys(user.MlKemPublicKey, user.EcdhPublicKey, user.EdPublicKey)
	if err != nil {
		return domain.InvalidData("invalid user public key")
	}

	isValid, err := crypto.VerifyEncryptedMasterKeySignature(publicKeys, key)
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		return domain.Failed("failed to verify signature")
	}
	if !isValid {
		return domain.InvalidData("invalid signature")
	}

	_, err = k.keys.Create(&domain.EncryptedKeys{
		UserID: user_id,
		Type:   "master",
		EncryptedKey: domain.EncryptedKey{
			Version:    crypto.EncodeBase64(key.Version),
			Ciphertext: crypto.EncodeBase64(key.Ciphertext),
			Nonce:      crypto.EncodeBase64(key.Nonce),
			Salt:       crypto.EncodeBase64(key.Salt),
			Signature:  crypto.EncodeBase64(key.Signature),
		},
	})
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		return domain.Failed("failed to save key")
	}

	return nil
}

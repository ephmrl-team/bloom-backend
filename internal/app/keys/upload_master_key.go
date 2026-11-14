package keys

import (
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (k *KeysApp) UploadMasterKey(user_id int, key *domain.EncryptedKeys) (*domain.EncryptedKeyBytes, error) {
	user, err := k.users.GetByID(user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("user not found")
		}
		logger.LogError(err.Error(), "keys-app")
		return nil, domain.Failed("failed to get user")
	}
	if user == nil {
		return nil, domain.NotFound("user not found")
	}

	_, err = k.keys.GetByUserID(user_id, key.Type)
	if err == nil {
		return nil, domain.AlreadyExists("key is already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		logger.LogError(err.Error(), "keys-app")
		return nil, domain.Failed("failed to get key")
	}

	if len(key.Ciphertext) > 10000 {
		return nil, domain.InvalidData("ciphertext too large")
	}
	ciphertextBytes, err := base64.StdEncoding.DecodeString(key.Ciphertext)
	if err != nil {
		return nil, domain.InvalidData("invalid ciphertext format")
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(key.Nonce)
	if err != nil || len(nonceBytes) != 12 {
		if err != nil {
			logger.LogError(err.Error(), "keys-app")
		}
		return nil, domain.InvalidData("invalid nonce")
	}

	saltBytes, err := base64.StdEncoding.DecodeString(key.Salt)
	if err != nil || len(saltBytes) != 32 {
		if err != nil {
			logger.LogError(err.Error(), "keys-app")
		}
		return nil, domain.InvalidData("invalid salt")
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(key.Signature)
	if err != nil || len(signatureBytes) != 114 {
		return nil, domain.InvalidData("invalid signature length")
	}

	edPublicKey, err := base64.StdEncoding.DecodeString(user.EdPublicKey)
	if err != nil || len(edPublicKey) != 57 {
		return nil, domain.InvalidData("invalid user public key")
	}

	isValid, err := crypto.VerifyEncryptedMasterKeySignature(edPublicKey, signatureBytes, key.Ciphertext, key.Nonce, key.Salt, user.MlKemPublicKey, user.EcdhPublicKey, user.EdPublicKey)
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		return nil, domain.Failed("failed to verify signature")
	}
	if !isValid {
		return nil, domain.InvalidData("invalid signature")
	}

	key, err = k.keys.Create(&domain.EncryptedKeys{
		UserID: user_id,
		Type:   key.Type,
		EncryptedKey: domain.EncryptedKey{
			Ciphertext: key.Ciphertext,
			Nonce:      key.Nonce,
			Salt:       key.Salt,
			Signature:  key.Signature,
		},
	})
	if err != nil {
		logger.LogError(err.Error(), "keys-app")
		return nil, domain.Failed("failed to save key")
	}

	return &domain.EncryptedKeyBytes{
		Ciphertext: ciphertextBytes,
		Nonce:      nonceBytes,
		Salt:       saltBytes,
		Signature:  signatureBytes,
	}, nil
}

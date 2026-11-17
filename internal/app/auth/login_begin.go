package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/generator"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (a *AuthApp) LoginBegin(authLookupID string) (*domain.KeysRequest, string, string, error) {
	user, err := a.users.GetByAuthLookupID(authLookupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", domain.NotFound("user is not found")
		}
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to get user")
	}
	if user == nil {
		return nil, "", "", domain.NotFound("user is not found")
	}

	ctx := context.Background()

	challenge, err := generator.GenerateChallenge()
	if err != nil {
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to generate challenge")
	}

	err = a.rdb.Set(ctx, "auth:challenge:"+user.PublicID, challenge, 2*time.Minute).Err()
	if err != nil {
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to save challenge")
	}

	encrypted_identity_keys, err := a.keysApp.GetUserKeys(user.ID, "bundle")
	if err != nil {
		return nil, "", "", err
	}

	encrypted_master_key, err := a.keysApp.GetUserKeys(user.ID, "master")
	if err != nil {
		return nil, "", "", err
	}

	publicKeys, err := crypto.DecodePublicKeys(user.MlKemPublicKey, user.EcdhPublicKey, user.EdPublicKey)
	if err != nil {
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to decode user public keys")
	}

	identityVersion, identityCiphertext, identityNonce, identitySalt, identitySignature, err := decodeEncryptedKey(encrypted_identity_keys.EncryptedKey)
	if err != nil {
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to decode identity keys")
	}

	masterVersion, masterCiphertext, masterNonce, masterSalt, masterSignature, err := decodeEncryptedKey(encrypted_master_key.EncryptedKey)
	if err != nil {
		logger.LogError(err.Error(), "auth-app")
		return nil, "", "", domain.Failed("failed to decode master key")
	}

	return &domain.KeysRequest{
		IdentityKeys: domain.IdentityKeysRequest{
			EncryptedSecretKeys: identity.EncryptedSecretKeys{
				Version:    identityVersion,
				Ciphertext: identityCiphertext,
				Nonce:      identityNonce,
				Salt:       identitySalt,
				Signature:  identitySignature,
			},
			IdentityPublicKeys: publicKeys,
		},
		EncryptedMasterKey: identity.EncryptedMasterKey{
			Version:    masterVersion,
			Ciphertext: masterCiphertext,
			Nonce:      masterNonce,
			Salt:       masterSalt,
			Signature:  masterSignature,
		},
	}, challenge, user.PublicID, nil
}

func decodeEncryptedKey(key domain.EncryptedKey) ([]byte, []byte, []byte, []byte, []byte, error) {
	version, err := crypto.DecodeBase64(key.Version)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	ciphertext, err := crypto.DecodeBase64(key.Ciphertext)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	nonce, err := crypto.DecodeBase64(key.Nonce)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	salt, err := crypto.DecodeBase64(key.Salt)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	signature, err := crypto.DecodeBase64(key.Signature)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return version, ciphertext, nonce, salt, signature, nil
}

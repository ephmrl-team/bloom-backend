package keys

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (k *KeysApp) GetUserKeys(user_id int, keys_type string) (*domain.EncryptedKeys, error) {
	keys, err := k.keys.GetByUserID(user_id, keys_type)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NotFound("keys not found")
		}
		logger.LogError(err.Error(), "keys-app")
		return nil, domain.Failed("failed to get keys")
	}

	if keys == nil {
		return nil, domain.NotFound("keys not found")
	}

	return keys, nil
}

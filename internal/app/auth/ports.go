package auth

import (
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

type SessionApp interface {
	CreateSession(user_id int) (string, *domain.Session, error)
}

type KeysApp interface {
	UploadIdentityKeys(user_id int, req *domain.IdentityKeysRequest) error
	UploadMasterKey(user_id int, key *identity.EncryptedMasterKey) error
	GetUserKeys(user_id int, keys_type string) (*domain.EncryptedKeys, error)
}

type UserRepo interface {
	GetByID(id int) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	GetByPublicID(id string) (*domain.User, error)
	GetByAuthLookupID(authLookupID string) (*domain.User, error)
	Create(user *domain.User) (*domain.User, error)
	Delete(id int) error
}

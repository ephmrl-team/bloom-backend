package auth

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func (a *AuthApp) Register(req *domain.KeysRequest) (string, *domain.User, *domain.Session, error) {
	publicKeys := req.IdentityKeys.IdentityPublicKeys

	publicID := identity.GenerateUserID(publicKeys.Ed448, publicKeys.X448, publicKeys.MlKem768)
	if err := (&identity.User{ID: publicID, PublicKeys: publicKeys}).Validate(); err != nil {
		return "", nil, nil, domain.InvalidData("invalid public keys")
	}

	user, err := a.users.GetByAuthLookupID(req.AuthLookupID)
	if user != nil {
		return "", nil, nil, domain.InvalidData("user with this auth lookup id is already exists")
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", nil, nil, domain.Failed("failed to create user")
	}

	createdUser, err := a.users.Create(&domain.User{
		AuthLookupID: req.AuthLookupID,
		PublicID:     publicID,
		Username:     publicID,
	})
	if err != nil {
		return "", nil, nil, domain.Failed("failed to create user")
	}

	rollback := func() {
		_ = a.users.Delete(createdUser.ID)
	}

	if err := a.keysApp.UploadIdentityKeys(createdUser.ID, &req.IdentityKeys); err != nil {
		rollback()
		return "", nil, nil, err
	}

	if err := a.keysApp.UploadMasterKey(createdUser.ID, &req.EncryptedMasterKey); err != nil {
		rollback()
		return "", nil, nil, err
	}

	finalUser, err := a.users.GetByID(createdUser.ID)
	if err != nil {
		return "", nil, nil, domain.Failed("failed to fetch registered user")
	}

	token, session, err := a.sessionApp.CreateSession(finalUser.ID)
	if err != nil {
		return "", nil, nil, err
	}

	return token, finalUser, session, nil
}

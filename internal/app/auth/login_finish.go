package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/pkg/crypto"
	"github.com/slipe-fun/skid-backend/internal/pkg/logger"
)

func (a *AuthApp) LoginFinish(user_id, signature string) (string, *domain.User, *domain.Session, error) {
	user, err := a.users.GetByPublicID(user_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, nil, domain.NotFound("user is not found")
		}
		logger.LogError(err.Error(), "auth-app")
		return "", nil, nil, domain.Failed("failed to get user")
	}
	if user == nil {
		return "", nil, nil, domain.NotFound("user is not found")
	}

	ctx := context.Background()

	challenge, err := a.rdb.GetDel(ctx, "auth:challenge:"+user_id).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil, nil, domain.Expired("login challenge expired or already used")
		}
		return "", nil, nil, domain.Failed("database error")
	}

	signatureBytes, err := crypto.DecodeBase64(signature)
	if err != nil || len(signatureBytes) != crypto.SignatureSize {
		return "", nil, nil, domain.InvalidData("invalid signature length")
	}

	edKeyBytes, err := crypto.DecodeBase64(user.EdPublicKey)
	if err != nil || len(edKeyBytes) != crypto.Ed448PublicKeySize {
		return "", nil, nil, domain.InvalidData("invalid public key length")
	}

	isValid, err := crypto.VerifyLoginChallengeSignature(edKeyBytes, signatureBytes, challenge, user_id)
	if err != nil {
		return "", nil, nil, domain.InvalidData("invalid signature")
	}

	if !isValid {
		return "", nil, nil, domain.InvalidData("invalid signature")
	}

	token, session, err := a.sessionApp.CreateSession(user.ID)
	if err != nil {
		return "", nil, nil, domain.Failed("failed to create session")
	}

	return token, user, session, nil
}

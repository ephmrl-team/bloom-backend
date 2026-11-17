package crypto

import (
	"errors"

	"github.com/cloudflare/circl/sign/ed448"
	"github.com/slipe-fun/skid-v4/pkg/identity"
	"github.com/slipe-fun/skid-v4/pkg/protocol"
)

func VerifyEncryptedMasterKeySignature(publicKeys identity.PublicKeys, encrypted *identity.EncryptedMasterKey) (bool, error) {
	if encrypted == nil {
		return false, errors.New("encrypted master key cannot be nil")
	}

	version, err := protocol.Parse(encrypted.Version)
	if err != nil {
		return false, err
	}

	if err := version.Check(); err != nil {
		return false, err
	}

	message := ConcatBytes(
		version.Bytes(),
		encrypted.Ciphertext,
		encrypted.Nonce,
		encrypted.Salt,
		publicKeys.MlKem768,
		publicKeys.X448,
		publicKeys.Ed448,
	)

	return ed448.Verify(publicKeys.Ed448, message, encrypted.Signature, version.Label("master_key")), nil
}

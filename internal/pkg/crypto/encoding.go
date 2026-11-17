package crypto

import (
	"encoding/base64"

	"github.com/slipe-fun/skid-v4/pkg/identity"
)

func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

func EncodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func DecodePublicKeys(mlKem, x448, ed448 string) (identity.PublicKeys, error) {
	mlKemBytes, err := DecodeBase64(mlKem)
	if err != nil {
		return identity.PublicKeys{}, err
	}

	x448Bytes, err := DecodeBase64(x448)
	if err != nil {
		return identity.PublicKeys{}, err
	}

	ed448Bytes, err := DecodeBase64(ed448)
	if err != nil {
		return identity.PublicKeys{}, err
	}

	return identity.PublicKeys{
		MlKem768: mlKemBytes,
		X448:     x448Bytes,
		Ed448:    ed448Bytes,
	}, nil
}

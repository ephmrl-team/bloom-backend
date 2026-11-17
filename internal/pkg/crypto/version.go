package crypto

import (
	"encoding/base64"

	"github.com/slipe-fun/skid-v4/pkg/protocol"
)

func ParseVersion(encoded string) (protocol.Version, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return protocol.Version{}, err
	}

	version, err := protocol.Parse(decoded)
	if err != nil {
		return protocol.Version{}, err
	}

	if err := version.Check(); err != nil {
		return protocol.Version{}, err
	}

	return version, nil
}

func ValidateVersion(encoded string) error {
	_, err := ParseVersion(encoded)
	return err
}

func ValidateVersionBytes(encoded []byte) error {
	version, err := protocol.Parse(encoded)
	if err != nil {
		return err
	}

	return version.Check()
}

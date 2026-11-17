package domain

import "github.com/slipe-fun/skid-v4/pkg/identity"

type EncryptedKey struct {
	Version    string `db:"version"`
	Ciphertext string `db:"ciphertext"`
	Nonce      string `db:"nonce"`
	Salt       string `db:"salt"`
	Signature  string `db:"signature"`
}

type EncryptedKeys struct {
	EncryptedKey
	ID     int    `db:"id" json:"id"`
	Type   string `db:"type" json:"type"`
	UserID int    `db:"user_id" json:"user_id"`
}

type IdentityKeysRequest struct {
	EncryptedSecretKeys identity.EncryptedSecretKeys `json:"encrypted_secret_keys"`
	IdentityPublicKeys  identity.PublicKeys          `json:"public_keys"`
}

type KeysRequest struct {
	AuthLookupID       string                      `json:"auth_lookup_id"`
	IdentityKeys       IdentityKeysRequest         `json:"identity_keys"`
	EncryptedMasterKey identity.EncryptedMasterKey `json:"encrypted_master_key"`
}

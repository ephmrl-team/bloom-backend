package domain

import (
	"encoding/json"
	"time"
)

type EncryptedSyncKey struct {
	CipherText string `json:"ciphertext" db:"ciphertext"`
	Nonce      string `json:"nonce" db:"nonce"`
}

type Handshake struct {
	Version             string `json:"version"`
	ReceiverCipherText  string `json:"receiver_ciphertext"`
	SenderEphemeralX448 string `json:"sender_ephemeral_x448"`
	SenderSignature     string `json:"sender_signature"`
	EncryptedSyncKey    `json:"encrypted_sync_key"`
}

func (h *Handshake) UnmarshalJSON(data []byte) error {
	type handshakeAlias Handshake
	aux := struct {
		*handshakeAlias
		LegacyReceiverCipherText string `json:"receiver_cipher_text"`
	}{
		handshakeAlias: (*handshakeAlias)(h),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if h.ReceiverCipherText == "" {
		h.ReceiverCipherText = aux.LegacyReceiverCipherText
	}

	return nil
}

type Chat struct {
	ID        int        `json:"id"`
	Members   []User     `json:"members"`
	Handshake *Handshake `json:"handshake,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type RawChat struct {
	Members   []Member   `json:"members"`
	Handshake *Handshake `json:"handshake"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

type Member struct {
	ID int `json:"id"`
}

type ChatWithLastMessage struct {
	ID              int        `json:"id"`
	Members         []User     `json:"members"`
	Handshake       *Handshake `json:"handshake,omitempty"`
	Type            string     `json:"type"`
	Title           string     `json:"title"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	LastMessage     *Message   `json:"last_message,omitempty"`
	LastReadMessage *Message   `json:"last_read_message,omitempty"`
}

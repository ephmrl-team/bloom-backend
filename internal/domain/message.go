package domain

import "time"

type Message struct {
	ID              int        `db:"id" json:"id"`
	Version         string     `db:"version" json:"version"`
	Ciphertext      string     `db:"ciphertext" json:"ciphertext"`
	Nonce           string     `db:"nonce" json:"nonce"`
	Salt            string     `db:"salt" json:"salt"`
	ChatID          int        `db:"chat_id" json:"chat_id"`
	Seen            *time.Time `db:"seen" json:"seen,omitempty"`
	ReplyTo         *int       `db:"reply_to" json:"reply_to,omitempty"`
	ClientMessageID *string    `db:"client_message_id" json:"client_message_id,omitempty"`
}

type MessageWithReply struct {
	Message
	ReplyToMessage *Message `json:"reply_to,omitempty"`
}

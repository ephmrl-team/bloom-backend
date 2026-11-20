package message

import (
	"database/sql"
	"errors"

	"github.com/slipe-fun/skid-backend/internal/domain"
)

func (r *MessageRepo) Create(message *domain.Message) (*domain.Message, bool, error) {
	query := `INSERT INTO messages
		(version, ciphertext, nonce, salt, chat_id, reply_to, client_message_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (chat_id, client_message_id) DO NOTHING
		RETURNING id, version, ciphertext, nonce, salt, chat_id, reply_to, client_message_id`

	created := domain.Message{}

	err := r.db.QueryRow(
		query,
		message.Version,
		message.Ciphertext,
		message.Nonce,
		message.Salt,
		message.ChatID,
		nullInt(message.ReplyTo),
		nullString(message.ClientMessageID),
	).Scan(
		&created.ID,
		&created.Version,
		&created.Ciphertext,
		&created.Nonce,
		&created.Salt,
		&created.ChatID,
		&created.ReplyTo,
		&created.ClientMessageID,
	)
	if err == nil {
		return &created, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}

	existing, err := r.getByChatAndClientMessageID(message.ChatID, message.ClientMessageID)
	if err != nil {
		return nil, false, err
	}

	return existing, true, nil
}

func (r *MessageRepo) getByChatAndClientMessageID(chatID int, clientMessageID *string) (*domain.Message, error) {
	query := `SELECT id, version, ciphertext, nonce, salt, chat_id, reply_to, client_message_id
		FROM messages
		WHERE chat_id = $1 AND client_message_id = $2`

	existing := domain.Message{}

	err := r.db.QueryRow(
		query,
		chatID,
		nullString(clientMessageID),
	).Scan(
		&existing.ID,
		&existing.Version,
		&existing.Ciphertext,
		&existing.Nonce,
		&existing.Salt,
		&existing.ChatID,
		&existing.ReplyTo,
		&existing.ClientMessageID,
	)
	if err != nil {
		return nil, err
	}

	return &existing, nil
}

func nullInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func nullString(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

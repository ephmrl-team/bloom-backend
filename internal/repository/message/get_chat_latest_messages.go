package message

import (
	"time"

	"github.com/slipe-fun/skid-backend/internal/domain"
	"github.com/slipe-fun/skid-backend/internal/metrics"
)

func (r *MessageRepo) GetChatLatestMessages(chatID, count int) ([]*domain.Message, error) {
	start := time.Now()

	rows, err := r.db.Query(`
	SELECT
		id,
		version,
		ciphertext,
		nonce,
		salt,
		chat_id,
		seen,
		reply_to
	FROM (
		SELECT
			id,
			version,
			ciphertext,
			nonce,
			salt,
			chat_id,
			seen,
			reply_to
		FROM messages
		WHERE chat_id = $1
		ORDER BY id DESC
		LIMIT $2
	) AS latest
	ORDER BY id ASC
	`, chatID, count)

	duration := time.Since(start)

	metrics.ObserveDB("messages_get_chat_latest", duration, err)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*domain.Message, 0)
	for rows.Next() {
		var message domain.Message
		err := rows.Scan(
			&message.ID,
			&message.Version,
			&message.Ciphertext,
			&message.Nonce,
			&message.Salt,
			&message.ChatID,
			&message.Seen,
			&message.ReplyTo,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &message)
	}
	return messages, rows.Err()
}

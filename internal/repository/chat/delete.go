package chat

import (
	"time"

	"github.com/slipe-fun/skid-backend/internal/metrics"
)

func (r *ChatRepo) Delete(id int) error {
	query := `DELETE FROM chats WHERE id = $1`

	start := time.Now()

	_, err := r.db.Exec(query, id)

	metrics.ObserveDB("chat_delete", time.Since(start), err)

	return err
}

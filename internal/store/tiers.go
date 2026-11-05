package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fleethacker/pb4t/internal/planner"
)

// migrateTiers persists guesses only for records without an assigned tier.
func (s *Store) migrateTiers(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"events", "templates"} {
		rows, err := tx.QueryContext(ctx, `SELECT id, payload FROM "`+table+`" WHERE json_type(payload, '$.tier') IS NULL OR json_type(payload, '$.tier') = 'null'`)
		if err != nil {
			return err
		}
		type rating struct {
			id   string
			tier int
		}
		var ratings []rating
		for rows.Next() {
			var id string
			var payload []byte
			if err := rows.Scan(&id, &payload); err != nil {
				rows.Close()
				return err
			}
			var record struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(payload, &record); err != nil {
				rows.Close()
				return err
			}
			ratings = append(ratings, rating{id: id, tier: planner.SuggestedTier(record.Type)})
		}
		err = rows.Err()
		closeErr := rows.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		for _, r := range ratings {
			if _, err := tx.ExecContext(ctx, `UPDATE "`+table+`" SET payload = json_set(payload, '$.tier', ?) WHERE id = ?`, r.tier, r.id); err != nil {
				return fmt.Errorf("assign %s tier: %w", table, err)
			}
		}
	}
	return tx.Commit()
}

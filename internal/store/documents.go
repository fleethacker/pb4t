package store

import (
	"context"
	"encoding/json"

	"github.com/fleethacker/pb4t/internal/planner"
)

func (s *Store) AddDocument(ctx context.Context, doc planner.TrainingDocument, original []byte) error {
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO documents(id,payload,original) VALUES(?,?,?)`, doc.ID, string(raw), original)
	return err
}
func (s *Store) Documents(ctx context.Context) ([]planner.TrainingDocument, error) {
	return read[planner.TrainingDocument](ctx, s.db, "documents")
}
func (s *Store) Document(ctx context.Context, id string) (planner.TrainingDocument, error) {
	var doc planner.TrainingDocument
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM documents WHERE id=?`, id).Scan(&raw)
	if err != nil {
		return doc, err
	}
	err = json.Unmarshal(raw, &doc)
	return doc, err
}
func (s *Store) DocumentFile(ctx context.Context, id string) ([]byte, error) {
	var original []byte
	err := s.db.QueryRowContext(ctx, `SELECT original FROM documents WHERE id=?`, id).Scan(&original)
	return original, err
}
func (s *Store) UpdateDocument(ctx context.Context, doc planner.TrainingDocument) error {
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE documents SET payload=? WHERE id=?`, string(raw), doc.ID)
	return err
}

// RecoverDocumentAnalyses makes interrupted imports retryable after a restart.
func (s *Store) RecoverDocumentAnalyses(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `UPDATE documents SET payload=json_set(payload,'$.status','failed','$.error','Analysis was interrupted by an app restart. The original PDF is saved; retry analysis.') WHERE json_extract(payload,'$.status')='analyzing'`)
	return err
}

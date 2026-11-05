package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fleethacker/pb4t/internal/planner"
	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

var tables = []string{"departments", "qualifications", "people", "duties", "assets", "constraints", "events", "templates", "requirements", "completions"}

func Open(path string, loc *time.Location, reset bool) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db}
	fail := func(err error) (*Store, error) { db.Close(); return nil, err }
	if _, err = db.Exec(`PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000;`); err != nil {
		return fail(err)
	}
	for _, table := range tables {
		if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS "` + table + `" (id TEXT PRIMARY KEY, payload TEXT NOT NULL)`); err != nil {
			return fail(err)
		}
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS person_qualifications (person_id TEXT REFERENCES people(id) ON DELETE CASCADE, qualification_id TEXT REFERENCES qualifications(id), PRIMARY KEY(person_id,qualification_id));
 CREATE TABLE IF NOT EXISTS event_constraints (event_id TEXT REFERENCES events(id) ON DELETE CASCADE, constraint_id TEXT REFERENCES "constraints"(id), PRIMARY KEY(event_id,constraint_id));
 CREATE TABLE IF NOT EXISTS assignments (event_id TEXT REFERENCES events(id) ON DELETE CASCADE, person_id TEXT REFERENCES people(id), role TEXT NOT NULL, PRIMARY KEY(event_id,person_id));
 CREATE TABLE IF NOT EXISTS proposals (id TEXT PRIMARY KEY, payload TEXT NOT NULL);
 CREATE TABLE IF NOT EXISTS documents (id TEXT PRIMARY KEY, payload TEXT NOT NULL, original BLOB NOT NULL);`)
	if err != nil {
		return fail(err)
	}
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM departments`).Scan(&count); err != nil {
		return fail(err)
	}
	if reset || count == 0 {
		if err = s.Save(context.Background(), planner.Seed(loc)); err != nil {
			return fail(err)
		}
		if reset {
			if _, err = db.Exec(`DELETE FROM proposals`); err != nil {
				return fail(err)
			}
		}
	}
	return s, nil
}
func (s *Store) Close() error { return s.db.Close() }
func read[T any](ctx context.Context, db *sql.DB, table string) ([]T, error) {
	order := "id"
	if table == "proposals" {
		order = "rowid"
	}
	rows, err := db.QueryContext(ctx, `SELECT payload FROM "`+table+`" ORDER BY `+order)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []T{}
	for rows.Next() {
		var raw []byte
		var item T
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) Load(ctx context.Context) (planner.Data, error) {
	var d planner.Data
	var err error
	if d.Departments, err = read[planner.Department](ctx, s.db, "departments"); err != nil {
		return d, err
	}
	if d.Qualifications, err = read[planner.Qualification](ctx, s.db, "qualifications"); err != nil {
		return d, err
	}
	if d.People, err = read[planner.Person](ctx, s.db, "people"); err != nil {
		return d, err
	}
	if d.Duties, err = read[planner.Duty](ctx, s.db, "duties"); err != nil {
		return d, err
	}
	if d.Assets, err = read[planner.Asset](ctx, s.db, "assets"); err != nil {
		return d, err
	}
	if d.Constraints, err = read[planner.Rule](ctx, s.db, "constraints"); err != nil {
		return d, err
	}
	if d.Events, err = read[planner.Event](ctx, s.db, "events"); err != nil {
		return d, err
	}
	if d.Templates, err = read[planner.EventTemplate](ctx, s.db, "templates"); err != nil {
		return d, err
	}
	if d.Requirements, err = read[planner.Requirement](ctx, s.db, "requirements"); err != nil {
		return d, err
	}
	if d.Completions, err = read[planner.Completion](ctx, s.db, "completions"); err != nil {
		return d, err
	}
	d.Assignments = []planner.Assignment{}
	rows, err := s.db.QueryContext(ctx, `SELECT event_id,person_id,role FROM assignments ORDER BY event_id,person_id`)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var a planner.Assignment
		if err = rows.Scan(&a.EventID, &a.PersonID, &a.Role); err != nil {
			return d, err
		}
		d.Assignments = append(d.Assignments, a)
	}
	return d, rows.Err()
}
func insert[T any](ctx context.Context, tx *sql.Tx, table string, items []T) error {
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return err
		}
		var key struct {
			ID string `json:"id"`
		}
		if err = json.Unmarshal(raw, &key); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO "`+table+`" (id,payload) VALUES (?,?)`, key.ID, string(raw)); err != nil {
			return fmt.Errorf("save %s: %w", table, err)
		}
	}
	return nil
}
func (s *Store) Save(ctx context.Context, d planner.Data) error { return s.save(ctx, d, nil) }

// SaveDocumentReview atomically creates the accepted planning record and marks
// its source suggestion as created, preventing duplicate acceptance on retry.
func (s *Store) SaveDocumentReview(ctx context.Context, d planner.Data, doc planner.TrainingDocument) error {
	return s.save(ctx, d, &doc)
}

func (s *Store) save(ctx context.Context, d planner.Data, doc *planner.TrainingDocument) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range []string{"assignments", "event_constraints", "person_qualifications"} {
		if _, err = tx.ExecContext(ctx, `DELETE FROM `+table); err != nil {
			return err
		}
	}
	for _, table := range tables {
		if _, err = tx.ExecContext(ctx, `DELETE FROM "`+table+`"`); err != nil {
			return err
		}
	}
	operations := []func() error{
		func() error { return insert(ctx, tx, "departments", d.Departments) }, func() error { return insert(ctx, tx, "qualifications", d.Qualifications) }, func() error { return insert(ctx, tx, "people", d.People) }, func() error { return insert(ctx, tx, "duties", d.Duties) }, func() error { return insert(ctx, tx, "assets", d.Assets) }, func() error { return insert(ctx, tx, "constraints", d.Constraints) }, func() error { return insert(ctx, tx, "events", d.Events) }, func() error { return insert(ctx, tx, "templates", d.Templates) }, func() error { return insert(ctx, tx, "requirements", d.Requirements) }, func() error { return insert(ctx, tx, "completions", d.Completions) }}
	for _, op := range operations {
		if err = op(); err != nil {
			return err
		}
	}
	for _, p := range d.People {
		for _, q := range p.Qualifications {
			if _, err = tx.ExecContext(ctx, `INSERT INTO person_qualifications VALUES (?,?)`, p.ID, q); err != nil {
				return err
			}
		}
	}
	for _, e := range d.Events {
		for _, id := range e.ConstraintIDs {
			if _, err = tx.ExecContext(ctx, `INSERT INTO event_constraints VALUES (?,?)`, e.ID, id); err != nil {
				return err
			}
		}
	}
	for _, a := range d.Assignments {
		if _, err = tx.ExecContext(ctx, `INSERT INTO assignments VALUES (?,?,?)`, a.EventID, a.PersonID, a.Role); err != nil {
			return err
		}
	}
	if doc != nil {
		raw, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE documents SET payload=? WHERE id=?`, string(raw), doc.ID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (s *Store) SaveProposal(ctx context.Context, p planner.Proposal) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO proposals VALUES (?,?) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload`, p.ID, string(raw))
	return err
}
func (s *Store) Proposal(ctx context.Context, id string) (planner.Proposal, error) {
	var raw []byte
	var p planner.Proposal
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM proposals WHERE id=?`, id).Scan(&raw)
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(raw, &p)
	return p, err
}
func (s *Store) Proposals(ctx context.Context) ([]planner.Proposal, error) {
	return read[planner.Proposal](ctx, s.db, "proposals")
}

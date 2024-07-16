package models

import (
	"database/sql"
)

type GlobalNote struct {
	ID       int    `json:"id"`
	Note     string `json:"note"`
	NoteDate string `json:"note_date"`
}

func (n *GlobalNote) Add(db *sql.DB) error {
	query := `INSERT INTO global_notes (note, note_date) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, n.Note, n.NoteDate).Scan(&n.ID)
	if err != nil {
		return err
	}
	return nil
}

func (n *GlobalNote) Get(db *sql.DB, noteDate string) error {
	query := `SELECT id, note, note_date FROM global_notes WHERE note_date = $1`
	return db.QueryRow(query, noteDate).Scan(&n.ID, &n.Note, &n.NoteDate)
}

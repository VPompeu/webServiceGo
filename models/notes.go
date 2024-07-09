package models

import (
	"database/sql"
)

type GlobalNote struct {
	ID       int    `json:"id"`
	Note     string `json:"note"`
	Sun      string `json:"sun"`
	Moon     string `json:"moon"`
	NoteDate string `json:"note_date"`
}

func (n *GlobalNote) Add(db *sql.DB) error {
	query := `INSERT INTO global_notes (note, sun, moon, note_date) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.QueryRow(query, n.Note, n.Sun, n.Moon, n.NoteDate).Scan(&n.ID)
	if err != nil {
		return err
	}
	return nil
}

func (n *GlobalNote) Get(db *sql.DB, noteDate string) error {
	query := `SELECT id, note, sun, moon, note_date FROM global_notes WHERE note_date = $1`
	return db.QueryRow(query, noteDate).Scan(&n.ID, &n.Note, &n.Sun, &n.Moon, &n.NoteDate)
}

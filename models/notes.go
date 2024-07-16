package models

import (
	"gorm.io/gorm"
)

type GlobalNote struct {
	ID       int    `json:"id"`
	Note     string `json:"note"`
	NoteDate string `json:"note_date"`
}

func (n *GlobalNote) Add(db *gorm.DB) error {
	result := db.Create(n)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (n *GlobalNote) Get(db *gorm.DB, noteDate string) error {
	result := db.Where("note_date = ?", noteDate).First(n)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

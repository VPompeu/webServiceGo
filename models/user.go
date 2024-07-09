package models

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Birthday string `json:"birthday"`
	City     string `json:"city"`
	State    string `json:"state"`
	Country  string `json:"country"`
	License  bool   `json:"license"`
}

type UserNote struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Note     string `json:"note"`
	NoteDate string `json:"note_date"`
}

func (u *User) GetUser(db *sql.DB) error {
	return db.QueryRow("SELECT name, email, password, phone, birthday, city, state, country, license FROM users WHERE id=$1",
		u.ID).Scan(&u.Name, &u.Email, &u.Password, &u.Phone, &u.Birthday, &u.City, &u.State, &u.Country, &u.License)
}

func (u *User) UpdateUser(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE users SET name=$1, email=$2, phone=$3, birthday=$4, city=$5, state=$6, country=$7, license=$8 WHERE id=$9",
			u.Name, u.Email, u.Phone, u.Birthday, u.City, u.State, u.Country, u.License, u.ID)

	return err
}

func (u *User) Activate(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE users SET license = true WHERE email=$1", u.Email)

	return err
}

func (u *User) DeleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE id=$1", u.ID)

	return err
}

func (u *User) CreateUser(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO users(name, email, password, phone, birthday, city, state, country, license) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id",
		u.Name, u.Email, u.Password, u.Phone, u.Birthday, u.City, u.State, u.Country, u.License).Scan(&u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) Login(db *sql.DB) (bool, error) {
	var userHash string

	err := db.QueryRow("SELECT id, password FROM users WHERE email = $1", u.Email).Scan(&u.ID, &userHash)

	if err == sql.ErrNoRows {
		return false, fmt.Errorf("usuário não encontrado")
	} else if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userHash), []byte(u.Password))
	if err != nil {
		return false, fmt.Errorf("email or password incorrect")
	}

	return true, nil
}

func (u *User) Register(db *sql.DB) error {
	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Query para inserir o novo usuário
	query := `INSERT INTO users (name, email, password, phone, birthday, city, state, country, license) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	err = db.QueryRow(query, u.Name, u.Email, hashedPassword, u.Phone, u.Birthday, u.City, u.State, u.Country, u.License).Scan(&u.ID)
	if err != nil {
		return err
	}

	return nil
}

func GetUsers(db *sql.DB, start, count int) ([]User, error) {
	rows, err := db.Query(
		"SELECT id, name, email, phone, birthday, city, state, country, license FROM users LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Birthday, &u.City, &u.State, &u.Country, &u.License); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (n *UserNote) AddUserNote(db *sql.DB) error {
	query := `INSERT INTO user_notes (user_id, note, note_date) VALUES ($1, $2, $3) RETURNING id`
	err := db.QueryRow(query, n.UserID, n.Note, n.NoteDate).Scan(&n.ID)
	if err != nil {
		return err
	}
	return nil
}

func (n *UserNote) GetUserNoteByDate(db *sql.DB, userID int, noteDate string) error {
	query := `SELECT id, user_id, note, note_date FROM user_notes WHERE user_id = $1 AND note_date = $2`
	return db.QueryRow(query, userID, noteDate).Scan(&n.ID, &n.UserID, &n.Note, &n.NoteDate)
}

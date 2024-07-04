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
	License  string `json:"license"`
}

func (u *User) GetUser(db *sql.DB) error {
	query := `SELECT name, email, password, phone, birthday, city, state, country, license FROM users WHERE id= ?`
	return db.QueryRow(query, u.ID).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Phone, &u.Birthday, &u.City, &u.State, &u.Country, &u.License)
}

func (u *User) UpdateUser(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE users SET name=?, email=?, phone=?, birthday=?, city=?, state=?, country=?, license=? WHERE id=?",
			u.Name, u.Email, u.Phone, u.Birthday, u.City, u.State, u.Country, u.License, u.ID)

	return err
}

func (u *User) DeleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE id=?", u.ID)

	return err
}

func (u *User) CreateUser(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO users(name, email, password, phone, birthday, city, state, country, license) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id",
		u.Name, u.Email, u.Password, u.Phone, u.Birthday, u.City, u.State, u.Country, u.License).Scan(&u.ID)

	if err != nil {
		return err
	}

	return nil
}

func (u *User) Login(db *sql.DB) (bool, error) {
	var userHash string

	err := db.QueryRow("SELECT password FROM users WHERE email = ?", u.Email).Scan(&userHash)

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

func GetUsers(db *sql.DB, start, count int) ([]User, error) {
	query := `
		SELECT id, name, email, phone, birthday, city, state, country, license FROM users LIMIT ?, ? `
	rows, err := db.Query(query, start, count)

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

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

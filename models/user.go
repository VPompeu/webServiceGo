package models

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

func (u *User) GetUser(db *gorm.DB) error {
	result := db.First(u, u.ID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) GetUserByEmail(db *gorm.DB) error {
	result := db.Where("email = ?", u.Email).First(u)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) UpdateUser(db *gorm.DB) error {
	result := db.Model(u).Updates(User{Name: u.Name, Email: u.Email, Phone: u.Phone, Birthday: u.Birthday, City: u.City, State: u.State, Country: u.Country, License: u.License})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) Activate(db *gorm.DB) error {
	result := db.Model(u).UpdateColumn("license", true)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) DeleteUser(db *gorm.DB) error {
	result := db.Delete(u, u.ID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) CreateUser(db *gorm.DB) error {
	u.Password = hashPassword(u.Password) // Hash da senha antes de salvar
	result := db.Create(u)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) Login(db *gorm.DB) (bool, error) {
	var user User
	result := db.Where("email = ?", u.Email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("usuário não encontrado")
		}
		return false, result.Error
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		return false, fmt.Errorf("email ou senha incorretos")
	}

	u.ID = user.ID // Atualiza o ID do usuário que fez login
	return true, nil
}

func (u *User) Register(db *gorm.DB) error {
	u.Password = hashPassword(u.Password) // Hash da senha antes de salvar
	result := db.Create(u)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) StoreResetToken(db *gorm.DB, token string) error {
	expiresAt := time.Now().Add(1 * time.Hour).UTC()
	resetToken := ResetToken{UserID: u.ID, Token: token, ExpiresAt: expiresAt}
	result := db.Create(&resetToken)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) ValidateResetToken(db *gorm.DB, token string) error {
	var resetToken ResetToken
	result := db.Where("token = ?", token).First(&resetToken)
	if result.Error != nil {
		return result.Error
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return fmt.Errorf("token expirado")
	}

	u.ID = resetToken.UserID // Atualiza o ID do usuário com base no token válido encontrado
	return nil
}

func (u *User) RemoveResetToken(db *gorm.DB, token string) error {
	result := db.Where("token = ?", token).Delete(&ResetToken{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *User) UpdateUserPassword(db *gorm.DB, newPassword string) error {
	hashedPassword := hashPassword(newPassword) // Hash da nova senha antes de atualizar
	result := db.Model(u).Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetUsers(db *gorm.DB, start, count int) ([]User, error) {
	var users []User
	result := db.Limit(count).Offset(start).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (n *UserNote) AddUserNote(db *gorm.DB) error {
	result := db.Create(n)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (n *UserNote) UpdateUserNote(db *gorm.DB) error {
	result := db.Model(n).Where("id = ? AND user_id = ?", n.ID, n.UserID).Updates(map[string]interface{}{"note": n.Note, "note_date": n.NoteDate})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (n *UserNote) GetUserNoteByDate(db *gorm.DB, userID int, noteDate string) error {
	result := db.Where("user_id = ? AND note_date = ?", userID, noteDate).First(n)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Função auxiliar para hashear a senha
func hashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

// Estrutura para tokens de redefinição de senha
type ResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    int       `gorm:"not null"`
	Token     string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

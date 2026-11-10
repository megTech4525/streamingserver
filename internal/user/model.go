package user

import "time"

type User struct {
	ID           int       `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password" json:"password"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// type Register struct {
// 	Username string `db:"username" json:"username"`
// 	Email    string `db:"email" json:"email"`
// 	Password string `db:"password" json:"password"`
// }

type registerReq struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginReq struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

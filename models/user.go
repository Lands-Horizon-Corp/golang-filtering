package models

import "time"

type User struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Age       int        `json:"age"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	Friend    UserFriend `json:"friend"`
}

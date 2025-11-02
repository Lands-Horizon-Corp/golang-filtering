package models

import "time"

type User struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Email     string       `json:"email"`
	Age       int          `json:"age"`
	Height    float64      `json:"height"`
	IsActive  bool         `json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	Birthday  time.Time    `json:"birthday"`
	Friend    UserFriend   `json:"friend"`
	Friends   []UserFriend `json:"friends"`
}

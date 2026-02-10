package domain

import "time"

type Event struct {
	ID     string    `json:"id"`
	UserID int       `json:"user_id"`
	Title  string    `json:"title"`
	Date   time.Time `json:"date"`
}
package domain

import "time"

type User struct {
	Id        int
	CreatedAt time.Time
	UpdatedAt time.Time
	Nickname  string
	ChatID    int64
	State     int
}

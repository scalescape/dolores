package cld

import "time"

type Object struct {
	Name      string    `json:"name"`
	Bucket    string    `json:"bucket"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

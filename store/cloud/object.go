package cloud

import "time"

type Object struct {
	Name    string    `json:"name"`
	Bucket  string    `json:"bucket"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

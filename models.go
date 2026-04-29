package models

import "time"

type Feedback struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Rating    int       `json:"rating"`
    Comments  string    `json:"comments,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}
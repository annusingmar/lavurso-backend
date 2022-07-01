package data

import "time"

type Thread struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Locked    bool      `json:"locked"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	UserID    int       `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
}

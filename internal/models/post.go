package models

import (
	"database/sql"
	"time"
)

type Post struct {
	ID        int64
	Content   string
	Tittle    string
	Tags      []string
	Media     sql.NullString
	CreatedAt time.Time
	UpdatedAt time.Time
	Comments  []Comment
	User      *User
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count,omitempty"`
}

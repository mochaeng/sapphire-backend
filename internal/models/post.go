package models

import "database/sql"

type Post struct {
	ID        int64
	Content   string
	Tittle    string
	Tags      []string
	Media     sql.NullString
	CreatedAt string
	UpdatedAt string
	Comments  []Comment
	User      *User
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}

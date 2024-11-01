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

type CreatePostPayload struct {
	Tittle  string   `json:"tittle" validate:"required,min=1,max=100"`
	Content string   `json:"content" validate:"required,min=1,max=1000"`
	Tags    []string `json:"tags,omitempty" validate:"max=5"`
}

type CreatePostResponse struct {
	ID        int64    `json:"id"`
	Tittle    string   `json:"tittle"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags,omitempty"`
	MediaURL  string   `json:"media_url,omitempty"`
	CreatedAt string   `json:"created_at"`
	UserID    int64    `json:"user_id"`
}

type UpdatePostPayload struct {
	Tittle  string `json:"tittle" validate:"omitempty,min=1,max=100"`
	Content string `json:"content" validate:"omitempty,min=1,max=1000"`
}

type UpdatePostResponse struct {
	Tittle    string `json:"tittle"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

type PostWithMetadata struct {
	Post
	CommentCount int `json:"comment_count"`
}

type GetPostResponse struct {
	Tittle    string       `json:"tittle"`
	Content   string       `json:"content"`
	Tags      []string     `json:"tags,omitempty"`
	MediaURL  string       `json:"media_url,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	User      UserResponse `json:"user"`
}

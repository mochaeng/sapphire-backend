package responses

import "time"

type CreatePostResponse struct {
	ID        int64     `json:"id"`
	Tittle    string    `json:"tittle"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	MediaURL  string    `json:"media_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UserID    int64     `json:"user_id"`
}

type UpdatePostResponse struct {
	Tittle    string    `json:"tittle"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostResponse struct {
	ID        int64         `json:"id"`
	Tittle    string        `json:"tittle,omitempty"`
	Content   string        `json:"content"`
	Tags      []string      `json:"tags,omitempty"`
	MediaURL  string        `json:"media_url,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	User      *UserResponse `json:"user,omitempty"`
}

type GetPostResponse struct {
	Tittle    string       `json:"tittle"`
	Content   string       `json:"content"`
	Tags      []string     `json:"tags,omitempty"`
	MediaURL  string       `json:"media_url,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	User      UserResponse `json:"user"`
}

type GetUserPostsResponse struct {
	Posts      []PostResponse `json:"posts"`
	User       *UserResponse  `json:"user"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

type FeedResponse struct {
	Posts      []PostResponse `json:"posts"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

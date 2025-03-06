package responses

type CreatePostResponse struct {
	ID        int64    `json:"id"`
	Tittle    string   `json:"tittle"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags,omitempty"`
	MediaURL  string   `json:"media_url,omitempty"`
	CreatedAt string   `json:"created_at"`
	UserID    int64    `json:"user_id"`
}

type UpdatePostResponse struct {
	Tittle    string `json:"tittle"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
}

type PostResponse struct {
	ID        int64         `json:"id"`
	Tittle    string        `json:"tittle,omitempty"`
	Content   string        `json:"content,omitempty"`
	Tags      []string      `json:"tags,omitempty"`
	MediaURL  string        `json:"media_url,omitempty"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
	User      *UserResponse `json:"user,omitempty"`
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

type GetUserPostsResponse struct {
	Posts []PostResponse `json:"posts"`
	User  *UserResponse  `json:"user"`
}

package responses

type UserResponse struct {
	ID        int64  `json:"id,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

type GetUserResponse struct {
	*UserResponse
}

type GetUserProfileResponse struct {
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name,omitempty"`
	Description   string `json:"description,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	BannerURL     string `json:"banner_url,omitempty"`
	Location      string `json:"location,omitempty"`
	UserLink      string `json:"user_link,omitempty"`
	NumFollowing  int    `json:"num_following"`
	NumFollowers  int    `json:"num_followers"`
	NumPosts      int    `json:"num_posts"`
	NumMediaPosts int    `json:"num_media_posts"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type RegisterUserResponse struct {
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	IsActive  bool   `json:"is_active"`
	Token     string `json:"token"`
}

type CreateTokenResponse struct {
	Token string `json:"token"`
}

type AuthMeResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username"`
	RoleName  string `json:"role_name"`
}

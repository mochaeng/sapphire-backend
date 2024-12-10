package payloads

type CreateUserPayload struct{}

type UpdateUserPayload struct{}

type DeleteUserPayload struct{}

type RegisterUserPayload struct {
	Username  string `json:"username" validate:"required,max=16,min=3"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=3,max=72"`
	FirstName string `json:"first_name" validate:"required,min=2,max=30"`
}

type SigninPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

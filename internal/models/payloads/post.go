package payloads

type CreatePostPayload struct {
	Tittle  string   `json:"tittle" validate:"required,min=1,max=100"`
	Content string   `json:"content" validate:"required,min=1,max=1000"`
	Tags    []string `json:"tags,omitempty" validate:"max=5"`
}

type UpdatePostPayload struct {
	Tittle  string `json:"tittle" validate:"omitempty,min=1,max=100"`
	Content string `json:"content" validate:"omitempty,min=1,max=1000"`
}

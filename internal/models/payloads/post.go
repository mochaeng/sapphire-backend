package payloads

type CreatePostDataValuesPayload struct {
	Tittle  string   `json:"tittle" validate:"max=100"`
	Content string   `json:"content" validate:"required,min=1,max=1000"`
	Tags    []string `json:"tags,omitempty" validate:"max=5"`
}

type UpdatePostPayload struct {
	Tittle  string `json:"tittle" validate:"omitempty,min=1,max=100"`
	Content string `json:"content" validate:"omitempty,min=1,max=1000"`
}

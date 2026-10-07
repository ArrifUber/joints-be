package dto

type CreateTranscriptRequest struct {
	Sequence int64  `json:"sequence" validate:"required,min=1"`
	Text     string `json:"text" validate:"required,min=1,max=5000"`
}

type CreateTranscriptResponse struct {
	Accepted bool  `json:"accepted"`
	Sequence int64 `json:"sequence"`
}


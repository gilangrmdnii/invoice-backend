package request

type CreateQCDocumentRequest struct {
	ProjectID    uint64 `json:"project_id" validate:"required,gt=0"`
	Title        string `json:"title" validate:"required,min=2,max=255"`
	Description  string `json:"description" validate:"omitempty,max=1000"`
	DocumentType string `json:"document_type" validate:"required,oneof=IMAGE AUDIO VIDEO DOCUMENT"`
	FileURL      string `json:"file_url" validate:"required"`
}

type UpdateQCDocumentRequest struct {
	Title        string `json:"title" validate:"omitempty,min=2,max=255"`
	Description  string `json:"description" validate:"omitempty,max=1000"`
	DocumentType string `json:"document_type" validate:"omitempty,oneof=IMAGE AUDIO VIDEO DOCUMENT"`
	FileURL      string `json:"file_url" validate:"omitempty"`
}

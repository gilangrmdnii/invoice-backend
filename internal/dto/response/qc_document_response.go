package response

import "time"

type QCDocumentResponse struct {
	ID           uint64    `json:"id"`
	ProjectID    uint64    `json:"project_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DocumentType string    `json:"document_type"`
	FileURL      string    `json:"file_url"`
	UploadedBy   uint64    `json:"uploaded_by"`
	UploaderName string    `json:"uploader_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

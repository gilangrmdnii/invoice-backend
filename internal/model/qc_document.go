package model

import "time"

type DocumentType string

const (
	DocTypeImage    DocumentType = "IMAGE"
	DocTypeAudio    DocumentType = "AUDIO"
	DocTypeVideo    DocumentType = "VIDEO"
	DocTypeDocument DocumentType = "DOCUMENT"
)

type QCDocument struct {
	ID           uint64       `json:"id"`
	ProjectID    uint64       `json:"project_id"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	DocumentType DocumentType `json:"document_type"`
	FileURL      string       `json:"file_url"`
	UploadedBy   uint64       `json:"uploaded_by"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

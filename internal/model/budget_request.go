package model

import "time"

type BudgetRequestStatus string

const (
	BudgetRequestPending  BudgetRequestStatus = "PENDING"
	BudgetRequestApproved BudgetRequestStatus = "APPROVED"
	BudgetRequestRejected BudgetRequestStatus = "REJECTED"
)

type BudgetRequest struct {
	ID               uint64              `json:"id"`
	ProjectID        uint64              `json:"project_id"`
	RequestedBy      uint64              `json:"requested_by"`
	Amount           float64             `json:"amount"`
	Reason           string              `json:"reason"`
	ProofURL         *string             `json:"proof_url,omitempty"`
	Status           BudgetRequestStatus `json:"status"`
	ApprovedBy       *uint64             `json:"approved_by,omitempty"`
	ApprovalNotes    *string             `json:"approval_notes,omitempty"`
	ApprovalProofURL *string             `json:"approval_proof_url,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

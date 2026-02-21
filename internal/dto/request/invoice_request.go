package request

type InvoiceItemRequest struct {
	Description string  `json:"description" validate:"required,min=1,max=500"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	Unit        string  `json:"unit" validate:"required,max=50"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

type InvoiceLabelRequest struct {
	Description string               `json:"description" validate:"required,min=1,max=500"`
	Items       []InvoiceItemRequest `json:"items" validate:"required,min=1,dive"`
}

type CreateInvoiceRequest struct {
	ProjectID        uint64               `json:"project_id" validate:"required"`
	InvoiceType      string               `json:"invoice_type" validate:"required,oneof=DP FINAL_PAYMENT TOP_1 TOP_2 TOP_3 MEALS ADDITIONAL"`
	RecipientName    string               `json:"recipient_name" validate:"required,max=255"`
	RecipientAddress string               `json:"recipient_address" validate:"max=1000"`
	Attention        string               `json:"attention" validate:"max=255"`
	PONumber         string               `json:"po_number" validate:"max=100"`
	InvoiceDate      string               `json:"invoice_date" validate:"required"`
	DPPercentage     *float64             `json:"dp_percentage" validate:"omitempty,gte=0,lte=100"`
	TaxPercentage    float64              `json:"tax_percentage" validate:"gte=0,lte=100"`
	Notes            string               `json:"notes" validate:"max=2000"`
	Language         string               `json:"language" validate:"required,oneof=ID EN"`
	FileURL          string               `json:"file_url" validate:"omitempty,max=500"`
	Items            []InvoiceItemRequest  `json:"items" validate:"omitempty,dive"`
	Labels           []InvoiceLabelRequest `json:"labels" validate:"omitempty,dive"`
}

type UpdateInvoiceRequest struct {
	RecipientName    string               `json:"recipient_name" validate:"omitempty,max=255"`
	RecipientAddress string               `json:"recipient_address" validate:"max=1000"`
	Attention        string               `json:"attention" validate:"max=255"`
	PONumber         string               `json:"po_number" validate:"max=100"`
	InvoiceDate      string               `json:"invoice_date"`
	DPPercentage     *float64             `json:"dp_percentage" validate:"omitempty,gte=0,lte=100"`
	TaxPercentage    *float64             `json:"tax_percentage" validate:"omitempty,gte=0,lte=100"`
	Notes            string               `json:"notes" validate:"max=2000"`
	Language         string               `json:"language" validate:"omitempty,oneof=ID EN"`
	FileURL          string               `json:"file_url" validate:"omitempty,max=500"`
	Items            []InvoiceItemRequest  `json:"items" validate:"omitempty,dive"`
	Labels           []InvoiceLabelRequest `json:"labels" validate:"omitempty,dive"`
}

type ApproveInvoiceRequest struct {
	Notes string `json:"notes" validate:"max=1000"`
}

type RejectInvoiceRequest struct {
	Notes string `json:"notes" validate:"required,min=5,max=1000"`
}

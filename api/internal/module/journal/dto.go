package journal

// CreateJournalRequest is the DTO payload for POST /journals.
type CreateJournalRequest struct {
	Date        string          `json:"date" example:"2026-04-21" validate:"required"`
	Description string          `json:"description,omitempty"`
	Lines       []JournalLineDto `json:"lines" validate:"required,min=2"`
}

// UpdateJournalRequest is the DTO payload for PUT /journals/{id}.
type UpdateJournalRequest struct {
	Date        string          `json:"date" example:"2026-04-21" validate:"required"`
	Description string          `json:"description,omitempty"`
	Lines       []JournalLineDto `json:"lines" validate:"required,min=2"`
}

// JournalLineDto is a single line in a journal entry request.
type JournalLineDto struct {
	AccountID int64   `json:"accountId" validate:"required"`
	Debit     float64 `json:"debit" validate:"min=0"`
	Credit    float64 `json:"credit" validate:"min=0"`
}

// JournalEntryResponse is the journal entry response DTO.
type JournalEntryResponse struct {
	ID          int64                  `json:"id"`
	Date        string                 `json:"date"`
	Description string                 `json:"description"`
	Lines       []JournalLineResponse  `json:"lines"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
}

// JournalLineResponse is a single line in a journal entry response.
type JournalLineResponse struct {
	ID             int64                 `json:"id"`
	JournalEntryID int64                 `json:"journalEntryId"`
	AccountID      int64                 `json:"accountId"`
	Debit          float64               `json:"debit"`
	Credit         float64               `json:"credit"`
	Account        *AccountInfoResponse  `json:"account,omitempty"`
}

// AccountInfoResponse contains basic account info embedded in journal lines.
type AccountInfoResponse struct {
	ID       int64  `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Level    int8   `json:"level"`
	IsPosting bool  `json:"isPosting"`
	ParentID *int64 `json:"parentId"`
}

// JournalSummary is the summary response for journals.
type JournalSummary struct {
	TotalDebit       float64 `json:"totalDebit"`
	TotalCredit      float64 `json:"totalCredit"`
	TransactionCount int64   `json:"transactionCount"`
}

// ListRequest is the query parameters for listing journal entries.
type ListRequest struct {
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	StartDate  string  `json:"startDate,omitempty"`
	EndDate    string  `json:"endDate,omitempty"`
	AccountIDs []int64 `json:"accountIds,omitempty"`
}

// ListResponse is the paginated list response for journal entries.
type ListResponse struct {
	Entries []JournalEntryResponse `json:"entries"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	Limit   int                    `json:"limit"`
}

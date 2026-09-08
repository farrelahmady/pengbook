package account

// CreateAccountRequest is the DTO payload for POST /accounts.
type CreateAccountRequest struct {
	Code     string `json:"code" example:"1.01.01.01" validate:"required"`
	Name     string `json:"name" example:"Mandiri - Main" validate:"required,min=1,max=200"`
	ParentID *int64 `json:"parentId,omitempty"`
}

// UpdateAccountRequest is the DTO payload for PUT /accounts/{id}.
type UpdateAccountRequest struct {
	Name     string `json:"name" example:"Mandiri - Main" validate:"required,min=1,max=200"`
	ParentID *int64 `json:"parentId,omitempty"`
}

// AccountResponse is the account response DTO.
type AccountResponse struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Level     int8   `json:"level"`
	IsPosting bool   `json:"isPosting"`
	ParentID  *int64 `json:"parentId"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// AccountWithChildren is a recursive tree structure for COA hierarchy.
type AccountWithChildren struct {
	ID        int64                  `json:"id"`
	Code      string                 `json:"code"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Level     int8                   `json:"level"`
	IsPosting bool                   `json:"isPosting"`
	ParentID  *int64                 `json:"parentId"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
	Children  []AccountWithChildren  `json:"children"`
}

// CoaTypeGroup groups accounts by type for the COA page.
type CoaTypeGroup struct {
	Type     string                `json:"type"`
	Label    string                `json:"label"`
	Icon     string                `json:"icon"`
	Accounts []AccountWithChildren `json:"accounts"`
	Count    int                   `json:"count"`
}

// CoaSummary is the summary response for the COA page.
type CoaSummary struct {
	TotalAccounts   int            `json:"totalAccounts"`
	PostingAccounts int            `json:"postingAccounts"`
	HeaderAccounts  int            `json:"headerAccounts"`
	Groups          []CoaTypeGroup `json:"groups"`
}

// PostingAccountResponse is a simplified response for posting accounts.
type PostingAccountResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

package account

// CreateAccountRequest is the DTO payload for POST /accounts.
// The code is generated server-side from the parent (max sibling segment + 1),
// so the client only sends the name and the parent. Level is derived from the
// parent (parent.level + 1) — the level/type columns are DB-generated.
type CreateAccountRequest struct {
	Name     string `json:"name" example:"Mandiri - Main" validate:"required,min=1,max=200"`
	ParentID *int64 `json:"parentId" validate:"required"`
}

// UpdateAccountRequest is the DTO payload for PUT /accounts/{id}.
// Rename-only: the code encodes the account position, so moving parents
// is not allowed here (that would need subtree re-coding).
type UpdateAccountRequest struct {
	Name string `json:"name" example:"Mandiri - Main" validate:"required,min=1,max=200"`
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

// AccountWithChildren is a recursive tree structure for Account hierarchy.
// Children uses pointers so buildTree links nodes instead of copying
// value snapshots (copying loses children attached later). JSON output
// is unchanged: an array of objects.
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
	Children  []*AccountWithChildren `json:"children"`
}

// AccountTypeGroup groups accounts by type for the Account page.
type AccountTypeGroup struct {
	Type         string                 `json:"type"`
	Label        string                 `json:"label"`
	Icon         string                 `json:"icon"`
	Accounts     []*AccountWithChildren `json:"accounts"`
	Count        int                    `json:"count"`
	PostingCount int                    `json:"postingCount"`
}

// AccountSummary is the lightweight summary response for the Account page.
// It carries only counts; the hierarchical tree lives in AccountTree.
type AccountSummary struct {
	TotalAccounts   int `json:"totalAccounts"`
	PostingAccounts int `json:"postingAccounts"`
	HeaderAccounts  int `json:"headerAccounts"`
}

// AccountTree is the hierarchical tree response for the Account page.
type AccountTree struct {
	Groups []AccountTypeGroup `json:"groups"`
}

// PostingAccountResponse is a simplified response for posting accounts.
type PostingAccountResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// ParentListItem is a lightweight candidate parent for the create-account form.
type ParentListItem struct {
	ID    int64  `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Level int8   `json:"level"`
}

type ParentsRequest struct {
	Level int8   `json:"level"`
	Type  string `json:"type"`
}

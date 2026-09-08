package account

import (
	"context"
	"errors"
	"time"

	"pengbook/api/internal/database"
)

// Errors
var (
	ErrNotFound      = errors.New("account not found")
	ErrCodeExists    = errors.New("account code already exists for this user")
	ErrInvalidCode   = errors.New("invalid account code format")
	ErrHasChildren   = errors.New("account has children, cannot delete")
	ErrHasJournalLines = errors.New("account has journal entry lines, cannot delete")
)

// Service is the PORT (interface) for account business logic.
type Service interface {
	// GetSummary returns the COA summary with hierarchical tree for a user.
	GetSummary(ctx context.Context, userID int64) (*CoaSummary, error)

	// Create creates a new account for the given user.
	Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error)

	// Update updates an existing account.
	Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error)

	// Delete deletes an account by id.
	Delete(ctx context.Context, userID int64, accountID int64) error

	// GetPostingAccounts returns all posting accounts (level=3) for a user.
	GetPostingAccounts(ctx context.Context, userID int64) ([]PostingAccountResponse, error)
}

type service struct {
	repo Repository
	tx   database.TxManager
}

func NewService(repo Repository, tx database.TxManager) Service {
	return &service{repo: repo, tx: tx}
}

func (s *service) GetSummary(ctx context.Context, userID int64) (*CoaSummary, error) {
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	posting, err := s.repo.CountPostingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	header, err := s.repo.CountHeaderByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accounts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build hierarchical tree
	tree := buildTree(accounts)

	// Group by type
	groups := groupByType(tree)

	return &CoaSummary{
		TotalAccounts:   int(total),
		PostingAccounts: int(posting),
		HeaderAccounts:  int(header),
		Groups:          groups,
	}, nil
}

func (s *service) Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error) {
	a := &Account{
		UserID:   userID,
		Code:     req.Code,
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	var created *Account
	err := s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, a); err != nil {
			return err
		}
		created = a

		// Audit log
		return s.repo.InsertAuditLog(ctx, &AccountAuditLog{
			UserID:    userID,
			AccountID: &a.ID,
			Action:    string(AccountActionCreated),
			NewValues: &JSONB{"code": a.Code, "name": a.Name, "type": a.Type, "level": a.Level},
		})
	})
	if err != nil {
		return nil, err
	}

	return toResponse(created), nil
}

func (s *service) Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error) {
	existing, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.UserID != userID {
		return nil, ErrNotFound
	}

	oldValues := JSONB{"name": existing.Name, "parent_id": existing.ParentID}

	existing.Name = req.Name
	existing.ParentID = req.ParentID

	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Update(ctx, existing); err != nil {
			return err
		}

		newValues := JSONB{"name": existing.Name, "parent_id": existing.ParentID}
		return s.repo.InsertAuditLog(ctx, &AccountAuditLog{
			UserID:    userID,
			AccountID: &accountID,
			Action:    string(AccountActionUpdated),
			OldValues: &oldValues,
			NewValues: &newValues,
		})
	})
	if err != nil {
		return nil, err
	}

	return toResponse(existing), nil
}

func (s *service) Delete(ctx context.Context, userID int64, accountID int64) error {
	existing, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if existing.UserID != userID {
		return ErrNotFound
	}

	return s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		// Audit log before delete
		if err := s.repo.InsertAuditLog(ctx, &AccountAuditLog{
			UserID:    userID,
			AccountID: &accountID,
			Action:    string(AccountActionDeleted),
			OldValues: &JSONB{"code": existing.Code, "name": existing.Name, "type": existing.Type, "level": existing.Level},
		}); err != nil {
			return err
		}

		return s.repo.Delete(ctx, accountID)
	})
}

func (s *service) GetPostingAccounts(ctx context.Context, userID int64) ([]PostingAccountResponse, error) {
	accounts, err := s.repo.FindPostingByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]PostingAccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = PostingAccountResponse{
			ID:   a.ID,
			Code: a.Code,
			Name: a.Name,
		}
	}
	return result, nil
}

// toResponse maps an Account entity to AccountResponse DTO.
func toResponse(a *Account) *AccountResponse {
	return &AccountResponse{
		ID:        a.ID,
		Code:      a.Code,
		Name:      a.Name,
		Type:      string(a.Type),
		Level:     a.Level,
		IsPosting: a.Level == 3,
		ParentID:  a.ParentID,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
}

// buildTree builds a hierarchical tree from a flat list of accounts.
func buildTree(accounts []Account) []AccountWithChildren {
	accountMap := make(map[int64]*AccountWithChildren)
	var roots []AccountWithChildren

	// First pass: create all nodes
	for _, a := range accounts {
		node := AccountWithChildren{
			ID:        a.ID,
			Code:      a.Code,
			Name:      a.Name,
			Type:      string(a.Type),
			Level:     a.Level,
			IsPosting: a.Level == 3,
			ParentID:  a.ParentID,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
			UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
			Children:  []AccountWithChildren{},
		}
		accountMap[a.ID] = &node
	}

	// Second pass: build tree
	for _, a := range accounts {
		node := accountMap[a.ID]
		if a.ParentID != nil {
			if parent, ok := accountMap[*a.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			} else {
				roots = append(roots, *node)
			}
		} else {
			roots = append(roots, *node)
		}
	}

	return roots
}

// groupByType groups accounts by their type for the COA page.
func groupByType(roots []AccountWithChildren) []CoaTypeGroup {
	typeOrder := []string{"ASSET", "LIABILITY", "EQUITY", "REVENUE", "EXPENSE", "OTHER"}
	typeLabels := map[string]string{
		"ASSET":     "Aset",
		"LIABILITY": "Liabilitas",
		"EQUITY":    "Ekuitas",
		"REVENUE":   "Pendapatan",
		"EXPENSE":   "Beban",
		"OTHER":     "Lain-lain",
	}
	typeIcons := map[string]string{
		"ASSET":     "wallet",
		"LIABILITY": "credit-card",
		"EQUITY":    "landmark",
		"REVENUE":   "trending-up",
		"EXPENSE":   "trending-down",
		"OTHER":     "wallet",
	}

	groupMap := make(map[string]*CoaTypeGroup)
	for _, t := range typeOrder {
		groupMap[t] = &CoaTypeGroup{
			Type:     t,
			Label:    typeLabels[t],
			Icon:     typeIcons[t],
			Accounts: []AccountWithChildren{},
			Count:    0,
		}
	}

	var countType func(nodes []AccountWithChildren) int
	countType = func(nodes []AccountWithChildren) int {
		count := 0
		for _, n := range nodes {
			count++
			count += countType(n.Children)
		}
		return count
	}

	for _, root := range roots {
		if group, ok := groupMap[root.Type]; ok {
			group.Accounts = append(group.Accounts, root)
			group.Count = countType(group.Accounts)
		}
	}

	// Remove empty groups
	var result []CoaTypeGroup
	for _, t := range typeOrder {
		if group := groupMap[t]; len(group.Accounts) > 0 {
			result = append(result, *group)
		}
	}

	return result
}

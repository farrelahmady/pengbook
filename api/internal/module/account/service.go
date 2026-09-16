package account

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"pengbook/api/internal/database"
	"pengbook/api/pkg/logger"
)

// Errors
var (
	ErrNotFound        = errors.New("account not found")
	ErrCodeExists      = errors.New("account code already exists for this user")
	ErrInvalidCode     = errors.New("invalid account code format")
	ErrHasChildren     = errors.New("account has children, cannot delete")
	ErrHasJournalLines = errors.New("account has journal entry lines, cannot delete")

	ErrParentRequired  = errors.New("parent account is required")
	ErrParentNotFound  = errors.New("parent account not found")
	ErrParentIsPosting = errors.New("posting accounts cannot have children")
	ErrCodeExhausted   = errors.New("no more codes available under this parent")
	ErrInvalidLevel    = errors.New("level must be between 0 and 2")
)

// Service is the PORT (interface) for account business logic.
type Service interface {
	// GetSummary returns the lightweight Account summary counts for a user.
	GetSummary(ctx context.Context, userID int64) (*AccountSummary, error)

	// GetTree returns the Account hierarchical tree grouped by type for a user.
	GetTree(ctx context.Context, userID int64) (*AccountTree, error)

	// GetParents returns candidate parents at the given level (0-2) for a user.
	GetParents(ctx context.Context, userID int64, level int8) ([]ParentListItem, error)

	// SeedRoots creates the six fixed level-0 roots for a user (idempotent).
	SeedRoots(ctx context.Context, userID int64) error

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

func (s *service) GetSummary(ctx context.Context, userID int64) (*AccountSummary, error) {
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

	return &AccountSummary{
		TotalAccounts:   int(total),
		PostingAccounts: int(posting),
		HeaderAccounts:  int(header),
	}, nil
}

func (s *service) GetTree(ctx context.Context, userID int64) (*AccountTree, error) {
	accounts, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build hierarchical tree
	tree := buildTree(accounts)

	// Group by type
	groups := groupByType(tree)

	return &AccountTree{
		Groups: groups,
	}, nil
}

func (s *service) Create(ctx context.Context, userID int64, req CreateAccountRequest) (*AccountResponse, error) {
	log := logger.FromContext(ctx)

	if req.ParentID == nil {
		return nil, ErrParentRequired
	}

	parent, err := s.repo.FindByID(ctx, *req.ParentID)
	if err != nil {
		log.Error("account create: failed to find parent", "user_id", userID, "parent_id", *req.ParentID, "error", err)
		return nil, err
	}
	if parent == nil || parent.UserID != userID {
		return nil, ErrParentNotFound
	}
	if parent.Level >= 3 {
		return nil, ErrParentIsPosting
	}

	code, err := s.nextChildCode(ctx, userID, parent)
	if err != nil {
		return nil, err
	}

	a := &Account{
		UserID:   userID,
		Code:     code,
		Name:     req.Name,
		ParentID: req.ParentID,
	}

	var created *Account
	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.repo.Create(ctx, a); err != nil {
			// Race: two concurrent creates generated the same code.
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return ErrCodeExists
			}
			return err
		}
		created = a

		// Asset posting accounts track a cached balance (starts at 0,
		// recalculated from journals). Other types have no balance row.
		if a.Type == AccountTypeAsset && a.Level == 3 {
			if err := s.repo.EnsureBalance(ctx, a.ID); err != nil {
				return err
			}
		}

		// Audit log
		return s.repo.InsertAuditLog(ctx, &AccountAuditLog{
			UserID:    userID,
			AccountID: &a.ID,
			Action:    string(AccountActionCreated),
			NewValues: &JSONB{"code": a.Code, "name": a.Name, "type": a.Type, "level": a.Level},
		})
	})
	if err != nil {
		log.Error("account create: failed to create account", "user_id", userID, "parent_id", *req.ParentID, "error", err)
		return nil, err
	}

	log.Info("account created", "user_id", userID, "account_id", a.ID, "code", a.Code, "name", a.Name)
	return toResponse(created), nil
}

// nextChildCode generates the next child code under the given parent:
// it copies the parent segments and bumps the segment at the child level
// to max(sibling segment) + 1, zeroing the remaining segments.
func (s *service) nextChildCode(ctx context.Context, userID int64, parent *Account) (string, error) {
	childLevel := parent.Level + 1 // 1..3

	maxCode, hasChildren, err := s.repo.FindMaxChildCode(ctx, userID, parent.ID)
	if err != nil {
		return "", err
	}

	parentSegs := strings.Split(parent.Code, ".")
	nextSeg := 1
	if hasChildren {
		childSegs := strings.Split(maxCode, ".")
		if len(childSegs) != 4 {
			return "", ErrInvalidCode
		}
		n, err := strconv.Atoi(childSegs[childLevel])
		if err != nil {
			return "", ErrInvalidCode
		}
		nextSeg = n + 1
	}
	if nextSeg > 99 {
		return "", ErrCodeExhausted
	}

	segs := make([]string, 4)
	copy(segs, parentSegs)
	segs[childLevel] = fmt.Sprintf("%02d", nextSeg)
	for i := int(childLevel) + 1; i < 4; i++ {
		segs[i] = "00"
	}
	return strings.Join(segs, "."), nil
}

func (s *service) GetParents(ctx context.Context, userID int64, level int8) ([]ParentListItem, error) {
	if level < 0 || level > 2 {
		return nil, ErrInvalidLevel
	}

	accounts, err := s.repo.FindByLevel(ctx, userID, level)
	if err != nil {
		return nil, err
	}

	result := make([]ParentListItem, len(accounts))
	for i, a := range accounts {
		result[i] = ParentListItem{
			ID:    a.ID,
			Code:  a.Code,
			Name:  a.Name,
			Type:  string(a.Type),
			Level: a.Level,
		}
	}
	return result, nil
}

func (s *service) SeedRoots(ctx context.Context, userID int64) error {
	log := logger.FromContext(ctx)
	if err := s.repo.SeedRoots(ctx, userID); err != nil {
		log.Error("account seed roots: failed", "user_id", userID, "error", err)
		return err
	}
	log.Info("account roots seeded", "user_id", userID)
	return nil
}

func (s *service) Update(ctx context.Context, userID int64, accountID int64, req UpdateAccountRequest) (*AccountResponse, error) {
	log := logger.FromContext(ctx)

	existing, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		log.Error("account update: failed to find account", "user_id", userID, "account_id", accountID, "error", err)
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
		log.Error("account update: failed to update account", "user_id", userID, "account_id", accountID, "error", err)
		return nil, err
	}

	log.Info("account updated", "user_id", userID, "account_id", accountID, "name", existing.Name)
	return toResponse(existing), nil
}

func (s *service) Delete(ctx context.Context, userID int64, accountID int64) error {
	log := logger.FromContext(ctx)

	existing, err := s.repo.FindByID(ctx, accountID)
	if err != nil {
		log.Error("account delete: failed to find account", "user_id", userID, "account_id", accountID, "error", err)
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if existing.UserID != userID {
		return ErrNotFound
	}

	err = s.tx.WithTransaction(ctx, func(ctx context.Context) error {
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
	if err != nil {
		log.Error("account delete: failed to delete account", "user_id", userID, "account_id", accountID, "error", err)
		return err
	}

	log.Info("account deleted", "user_id", userID, "account_id", accountID, "code", existing.Code, "name", existing.Name)
	return nil
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
// Nodes are linked by pointer so children attached later are visible
// through parents/roots already collected, regardless of input order.
func buildTree(accounts []Account) []*AccountWithChildren {
	accountMap := make(map[int64]*AccountWithChildren, len(accounts))
	var roots []*AccountWithChildren

	// First pass: create all nodes
	for _, a := range accounts {
		node := &AccountWithChildren{
			ID:        a.ID,
			Code:      a.Code,
			Name:      a.Name,
			Type:      string(a.Type),
			Level:     a.Level,
			IsPosting: a.Level == 3,
			ParentID:  a.ParentID,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
			UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
			Children:  []*AccountWithChildren{},
		}
		accountMap[a.ID] = node
	}

	// Second pass: link nodes (no value copies)
	for _, a := range accounts {
		node := accountMap[a.ID]
		if a.ParentID != nil {
			if parent, ok := accountMap[*a.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	return roots
}

// groupByType groups accounts by their type for the Account page.
func groupByType(roots []*AccountWithChildren) []AccountTypeGroup {
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

	groupMap := make(map[string]*AccountTypeGroup)
	for _, t := range typeOrder {
		groupMap[t] = &AccountTypeGroup{
			Type:         t,
			Label:        typeLabels[t],
			Icon:         typeIcons[t],
			Accounts:     []*AccountWithChildren{},
			Count:        0,
			PostingCount: 0,
		}
	}

	// countSubtree returns (total, posting) for a subtree in a single pass.
	var countSubtree func(nodes []*AccountWithChildren) (total, posting int)
	countSubtree = func(nodes []*AccountWithChildren) (total, posting int) {
		for _, n := range nodes {
			total++
			if n.IsPosting {
				posting++
			}
			childTotal, childPosting := countSubtree(n.Children)
			total += childTotal
			posting += childPosting
		}
		return total, posting
	}

	for _, root := range roots {
		if group, ok := groupMap[root.Type]; ok {
			group.Accounts = append(group.Accounts, root)
			group.Count, group.PostingCount = countSubtree(group.Accounts)
		}
	}

	// Remove empty groups
	var result []AccountTypeGroup
	for _, t := range typeOrder {
		if group := groupMap[t]; len(group.Accounts) > 0 {
			result = append(result, *group)
		}
	}

	return result
}

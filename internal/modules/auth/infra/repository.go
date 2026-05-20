// Package infra provides auth infrastructure implementations.
package infra

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"github.com/masmuss/gokit-starter/internal/database/ent"
	entuser "github.com/masmuss/gokit-starter/internal/database/ent/user"
	"github.com/masmuss/gokit-starter/internal/infra/database"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

const (
	organizationTypeCompany   = "company"
	organizationTypePersonal  = "personal"
	organizationStatusActive  = "active"
	userStatusActive          = "active"
	organizationCodeMaxLength = 16
)

// Repository persists auth entities using Ent.
type Repository struct {
	client *ent.Client
}

// NewRepository creates a new auth repository.
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// NewRepositoryFromDB creates a Repository from database.DB.
func NewRepositoryFromDB(db *database.DB) *Repository {
	return NewRepository(db.Client)
}

// CreateAccount creates an organization and its first user in one transaction.
func (r *Repository) CreateAccount(
	ctx context.Context,
	input domain.RegisterInput,
	passwordHash string,
) (result domain.User, err error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}

		err = tx.Commit()
	}()

	// Handle personal vs company account
	orgName := input.OrganizationName
	orgType := organizationTypeCompany
	if orgName == "" {
		orgName = input.Name
		orgType = organizationTypePersonal
	}

	// Try creating organization with unique code (retry on collision)
	var org *ent.Organization
	for i := 0; i < 3; i++ {
		org, err = tx.Organization.Create().
			SetName(orgName).
			SetCode(organizationCode(orgName)).
			SetType(orgType).
			SetStatus(organizationStatusActive).
			Save(ctx)

		if err == nil {
			break
		}

		if !ent.IsConstraintError(err) {
			return domain.User{}, fmt.Errorf("create organization: %w", err)
		}
	}

	if err != nil {
		return domain.User{}, fmt.Errorf("create organization: %w", err)
	}

	userRecord, err := tx.User.Create().
		SetOrganizationID(org.ID).
		SetName(input.Name).
		SetEmail(strings.ToLower(input.Email)).
		SetPasswordHash(passwordHash).
		SetStatus(userStatusActive).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return domain.User{}, domain.ErrEmailAlreadyUsed
		}

		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return toDomainUser(userRecord, org), nil
}

// FindByEmail returns a user with its organization by email.
func (r *Repository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	userRecord, err := r.client.User.Query().
		Where(entuser.EmailEQ(strings.ToLower(strings.TrimSpace(email)))).
		WithOrganization().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return toDomainUser(userRecord, userRecord.Edges.Organization), nil
}

// FindByID returns a user with its organization by ID.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	userRecord, err := r.client.User.Query().
		Where(entuser.IDEQ(id)).
		WithOrganization().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return domain.User{}, domain.ErrUserNotFound
		}

		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return toDomainUser(userRecord, userRecord.Edges.Organization), nil
}

func toDomainUser(userRecord *ent.User, orgRecord *ent.Organization) domain.User {
	return domain.User{
		ID:           userRecord.ID,
		Name:         userRecord.Name,
		Email:        userRecord.Email,
		PasswordHash: userRecord.PasswordHash,
		Status:       string(userRecord.Status),
		Organization: domain.Organization{
			ID:     orgRecord.ID,
			Name:   orgRecord.Name,
			Code:   orgRecord.Code,
			Type:   orgRecord.Type,
			Status: orgRecord.Status,
		},
	}
}

func organizationCode(name string) string {
	base := normalizeCode(name)
	if base == "" {
		base = "org"
	}

	// Use a longer suffix (8 chars) and better truncation
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	maxBaseLength := organizationCodeMaxLength - len(suffix) - 1
	if maxBaseLength < 1 {
		maxBaseLength = 1
	}
	if len(base) > maxBaseLength {
		base = base[:maxBaseLength]
	}

	return fmt.Sprintf("%s-%s", base, suffix)
}

func normalizeCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	return strings.Trim(strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		case unicode.IsSpace(r), r == '-', r == '_':
			return '-'
		default:
			return -1
		}
	}, value), "-")
}

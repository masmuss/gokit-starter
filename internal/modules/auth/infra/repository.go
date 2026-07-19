// Package infra provides auth infrastructure implementations.
package infra

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/masmuss/gokit-starter/internal/database/model"
	"github.com/masmuss/gokit-starter/internal/infra/database"
	"github.com/masmuss/gokit-starter/internal/modules/auth/domain"
)

const (
	organizationTypeCompany  = "company"
	organizationTypePersonal = "personal"
	roleAdmin                = "admin"
	organizationCodeMaxLen   = 16
)

// Repository persists auth entities using GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// NewRepositoryFromDB creates a Repository from database.DB.
func NewRepositoryFromDB(database *database.DB) *Repository {
	return NewRepository(database.DB)
}

// CreateAccount creates an organization and its first user in one transaction.
func (r *Repository) CreateAccount(
	ctx context.Context,
	input domain.RegisterInput,
	passwordHash string,
) (domain.User, error) {
	orgName := input.OrganizationName
	orgType := organizationTypeCompany
	if orgName == "" {
		orgName = input.Name
		orgType = organizationTypePersonal
	}

	var result domain.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var org model.Organization

		for i := 0; i < 3; i++ {
			org = model.Organization{
				ID:   uuid.New(),
				Name: orgName,
				Code: organizationCode(orgName),
				Type: orgType,
			}

			createErr := tx.Create(&org).Error
			if createErr == nil {
				break
			}
			if !isUniqueViolation(createErr) {
				return fmt.Errorf("create organization: %w", createErr)
			}
		}

		userRecord := model.User{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			Name:           input.Name,
			Email:          strings.ToLower(input.Email),
			PasswordHash:   passwordHash,
			Role:           roleAdmin,
		}

		if err := tx.Create(&userRecord).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrEmailAlreadyUsed
			}
			return fmt.Errorf("create user: %w", err)
		}

		userRecord.Organization = org
		result = toDomainUser(&userRecord)
		return nil
	})

	return result, err
}

// FindByEmail returns a user with its organization by email.
func (r *Repository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var record model.User
	err := r.db.WithContext(ctx).
		Preload("Organization").
		Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}

	return toDomainUser(&record), nil
}

// FindByID returns a user with its organization by ID.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var record model.User
	err := r.db.WithContext(ctx).
		Preload("Organization").
		First(&record, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("find user by id: %w", err)
	}

	return toDomainUser(&record), nil
}

// UpdatePassword sets a new password hash for the user.
func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func toDomainUser(record *model.User) domain.User {
	return domain.User{
		ID:           record.ID,
		Name:         record.Name,
		Email:        record.Email,
		PasswordHash: record.PasswordHash,
		Status:       record.Status,
		Role:         record.Role,
		Organization: domain.Organization{
			ID:   record.Organization.ID,
			Name: record.Organization.Name,
			Code: record.Organization.Code,
			Type: record.Organization.Type,
		},
	}
}

func organizationCode(name string) string {
	base := normalizeCode(name)
	if base == "" {
		base = "org"
	}

	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	maxBase := organizationCodeMaxLen - len(suffix) - 1
	if maxBase < 1 {
		maxBase = 1
	}
	if len(base) > maxBase {
		base = base[:maxBase]
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

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint")
}

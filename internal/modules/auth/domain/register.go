package domain

// RegisterInput contains data required to create a new account.
type RegisterInput struct {
	Name             string
	Email            string
	Password         string
	OrganizationName string // Optional for personal accounts
}

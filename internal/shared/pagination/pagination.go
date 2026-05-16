// Package pagination provides helpers for request pagination.
package pagination

// Params represents pagination parameters.
type Params struct {
	Page    int `json:"page"     query:"page"`
	PerPage int `json:"per_page" query:"per_page"`
}

// NewParams creates pagination params with defaults.
func NewParams(page, perPage int) Params {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if perPage > 100 {
		perPage = 100
	}

	return Params{
		Page:    page,
		PerPage: perPage,
	}
}

// Offset returns the SQL offset.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL limit.
func (p Params) Limit() int {
	return p.PerPage
}

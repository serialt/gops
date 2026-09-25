package utils

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 1000
)

func PageBounds(page, limit int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	return (page - 1) * limit, limit
}

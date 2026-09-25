package api

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 1000
)

type PageRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (r *PageRequest) Normalize() {
	if r.Page < 1 {
		r.Page = DefaultPage
	}
	if r.Limit < 1 {
		r.Limit = DefaultLimit
	}
	if r.Limit > MaxLimit {
		r.Limit = MaxLimit
	}
}

func (r *PageRequest) Offset() int { r.Normalize(); return (r.Page - 1) * r.Limit }

type PageResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

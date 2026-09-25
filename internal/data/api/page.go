package api

const MaxLimit = 1000

type PageRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PageResponse[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
}

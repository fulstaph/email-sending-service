package models

type PaginationParams struct {
	Page    int `form:"page,default=1"`
	PerPage int `form:"per_page,default=20"`
}

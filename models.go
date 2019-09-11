package infra

import "gopkg.in/go-playground/validator.v9"

type Pagination struct {
	page, pageSize uint64
	Count          uint64
	Data           interface{}
}

func NewPagination(page, pageSize uint64) *Pagination {
	return &Pagination{page: page, pageSize: pageSize, Data: make([]interface{}, 0)}
}

func (p Pagination) Offset() uint64 {
	return (p.page - 1) * p.pageSize
}

func (p Pagination) Limit() uint64 {
	return p.pageSize
}

type IDRequest struct {
	ID uint64 `validate:"required"`
}

func (req *IDRequest) Validate() error {
	return validator.New().Struct(req)
}

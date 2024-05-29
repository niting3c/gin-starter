package utils

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Limit      int         `json:"limit,omitempty;query:limit"`
	Page       int         `json:"page,omitempty;query:page"`
	Sort       string      `json:"sort,omitempty;query:sort"`
	TotalRows  int64       `json:"total_rows"`
	TotalPages int         `json:"total_pages"`
	Rows       interface{} `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}
func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}
func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}
func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "\"id\" desc"
	}
	return p.Sort
}

// PaginateQueryExtractor extracts pagination query parameters from a Gin context
// and returns a config.Pagination object along with an error message if any.
func PaginateQueryExtractor(context *gin.Context, validSortFields []string) (*Pagination, *ErrorMessage) {
	page, err := strconv.Atoi(context.Query("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(context.Query("per_page"))
	if err != nil || pageSize <= 0 {
		pageSize = 10 // Default to 10 if not specified or invalid
	}
	sort := context.Query("sort")
	direction := context.Query("sortDesc")
	var sortWithDirection string
	if sort != "" {
		// Validate sort field
		if !StringContains(validSortFields, sort) {
			return nil, &ErrorMessage{StatusCode: http.StatusBadRequest, Message: "Invalid sort field"}
		}
		if direction == "true" {
			sortWithDirection = fmt.Sprintf("\"%s\" DESC", sort)
		} else {
			sortWithDirection = fmt.Sprintf("\"%s\" ASC", sort)
		}
	}

	return &Pagination{
		Limit: pageSize,
		Page:  page,
		Sort:  sortWithDirection,
	}, nil
}

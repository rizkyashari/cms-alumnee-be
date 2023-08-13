package util

import (
	"strings"

	"gorm.io/gorm"
)

func Pagination(Limit, Page int, SortBy, SortOrder string, validColumnName []string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sortBy := "created_at"
		if SortBy != "" {
			paramSortBy := strings.ToLower(SortBy)
			for _, valid := range validColumnName {
				if paramSortBy == valid {
					sortBy = paramSortBy
				}
			}
		}

		sortOrder := "DESC"
		validSortOrder := []string{"ASC", "DESC"}
		if SortOrder != "" {
			paramSortOrder := strings.ToUpper(SortOrder)
			for _, valid := range validSortOrder {
				if paramSortOrder == valid {
					sortOrder = paramSortOrder
				}
			}
		}

		limit := 10
		if Limit > 0 {
			limit = Limit
		}

		page := 1
		if Page > 0 {
			page = Page
		}
		offset := (page - 1) * limit

		if Limit == -1 {
			limit = 999
			offset = 0
		}

		return db.Order((sortBy + " " + sortOrder)).Offset(offset).Limit(limit)
	}
}

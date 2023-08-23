package util

import (
	"strconv"
	"strings"

	"github.com/fadhln/lms-be/delivery/rq"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
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

func GetMaxPage(db *gorm.DB, limit int) int {
	count := int(db.RowsAffected)
	maxPage := count / limit
	if (count % limit) > 0 {
		maxPage += 1
	}

	return maxPage
}

func IsParamValid(params *rq.PaginationParams[interface{}]) bool {
	if params.Limit <= 0 {
		return false
	}

	if params.Page <= 0 {
		return false
	}

	if len(params.SortOrder) > 0 {
		if !((params.SortOrder == "ASC") || (params.SortOrder == "DESC")) {
			return false
		}
	}

	return true
}

func ParseQuery(c *gin.Context) (limit int, page int, sortBy string, sortOrder string, err error) {
	limitQuery := c.DefaultQuery("limit", "10")
	limitInt, err := strconv.Atoi(limitQuery)
	if err != nil {
		return 0, 0, "", "", errmsg.ErrRequestParamsInvalid
	}
	pageQuery := c.DefaultQuery("page", "1")
	pageInt, err := strconv.Atoi(pageQuery)
	if err != nil {
		return 0, 0, "", "", errmsg.ErrRequestParamsInvalid
	}
	sortByQuery := c.DefaultQuery("sortby", "created_at")
	sortOrderQuery := c.DefaultQuery("order", "desc")

	return limitInt, pageInt, sortByQuery, strings.ToUpper(sortOrderQuery), nil
}

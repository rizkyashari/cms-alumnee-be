package util

import (
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

func GetAccountContext(c *gin.Context) (*rs.AccountResponse, error) {
	var account rs.AccountResponse
	gotAccount, exist := c.Get("account")

	if !exist {
		return nil, &errmsg.ErrIsEmpty{FieldName: "User"}
	}
	err := copier.Copy(&account, gotAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	return &account, nil
}

package util

import (
	"github.com/fadhln/lms-be/constants"
	"github.com/fadhln/lms-be/delivery/rs"
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

func GetAccountContext(c *gin.Context, expectAccType int) (*rs.AccountResponse, error) {
	var account rs.AccountResponse
	gotAccount, exist := c.Get("account")

	if !exist {
		return nil, &errmsg.ErrIsEmpty{FieldName: "User"}
	}
	err := copier.Copy(&account, gotAccount)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if expectAccType == constants.ACCOUNT_STUDENT {
		if account.Student == nil || account.AccountType != constants.ACCOUNT_STUDENT {
			return nil, &errmsg.ErrUserIsNot{FieldName: "Student"}
		}
	} else if expectAccType == constants.ACCOUNT_TEACHER {
		if account.Teacher == nil || account.AccountType != constants.ACCOUNT_TEACHER {
			return nil, &errmsg.ErrUserIsNot{FieldName: "Teacher"}
		}
	} else if expectAccType == constants.ACCOUNT_ADMIN {
		if account.AccountType != constants.ACCOUNT_ADMIN {
			return nil, &errmsg.ErrUserIsNot{FieldName: "Admin"}
		}
	}

	return &account, nil
}

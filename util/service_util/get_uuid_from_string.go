package serviceutil

import (
	"github.com/fadhln/lms-be/util/errmsg"
	"github.com/google/uuid"
)

func GetUUIDFromStringWithValidation(idName string, strId *string) (*uuid.UUID, error) {
	if strId == nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: idName}
	}

	if len(*strId) != 36 {
		return nil, &errmsg.ErrFieldIsWrong{FieldName: idName}
	}

	parsedId, err := uuid.Parse(*strId)
	if err != nil {
		return nil, &errmsg.ErrInternal{Err: err}
	}

	if parsedId == uuid.Nil {
		return nil, &errmsg.ErrIsEmpty{FieldName: idName}
	}

	return &parsedId, nil
}

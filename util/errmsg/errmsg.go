package errmsg

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRequestBodyInvalid   = errors.New("request body is invalid")
	ErrRequestHeaderInvalid = errors.New("request header is invalid")
	ErrRequestParamsInvalid = errors.New("request parameter is invalid")
)

type ErrUserIsNot struct {
	FieldName string
}

func (e *ErrUserIsNot) Error() string {
	return fmt.Sprintf("user is not %s", e.FieldName)
}

type ErrAlreadyUsed struct {
	FieldName string
}

func (e *ErrAlreadyUsed) Error() string {
	return fmt.Sprintf("%s is already used", e.FieldName)
}

type ErrAlreadyExist struct {
	FieldName string
}

func (e *ErrAlreadyExist) Error() string {
	return fmt.Sprintf("%s is already exist", e.FieldName)
}

type ErrFieldIsWrong struct {
	FieldName string
}

func (e *ErrFieldIsWrong) Error() string {
	return fmt.Sprintf("%s is wrong", e.FieldName)
}

type ErrFieldIsExpired struct {
	FieldName string
}

func (e *ErrFieldIsExpired) Error() string {
	return fmt.Sprintf("%s is expired", e.FieldName)
}

type ErrNotFound struct {
	FieldName string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s is not found", e.FieldName)
}

type ErrIsEmpty struct {
	FieldName string
}

func (e *ErrIsEmpty) Error() string {
	return fmt.Sprintf("%s is empty", e.FieldName)
}

type ErrMinAmount struct {
	Amount int
}

func (e *ErrMinAmount) Error() string {
	return fmt.Sprintf("a minimum of %d is not meet", e.Amount)
}

type ErrMaxAmount struct {
	Amount int
}

func (e *ErrMaxAmount) Error() string {
	return fmt.Sprintf("a maximum of %d is exceeded", e.Amount)
}

type ErrInvalidDesc struct {
	FieldName string
}

func (e *ErrInvalidDesc) Error() string {
	return fmt.Sprintf("description provided is invalid. %s", e.FieldName)
}

type ErrInternal struct {
	Err error
}

func (e *ErrInternal) Error() string {
	if strings.Contains(e.Err.Error(), "23505") {
		return fmt.Sprintf("Internal Error: %s", "Duplicate key")
	}

	return fmt.Sprintf("Internal Error: %s", e.Err.Error())
}

type ErrANotSameB struct {
	A string
	B string
}

func (e *ErrANotSameB) Error() string {
	return fmt.Sprintf("%s is not the same as %s", e.A, e.B)
}

type ErrRequestFileInvalid struct {
	Info string
}

func (e *ErrRequestFileInvalid) Error() string {
	return fmt.Sprintf("Request file is invalid. Info: %s", e.Info)
}

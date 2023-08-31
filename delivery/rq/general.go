package rq

import "mime/multipart"

type IDOnlyRequest struct {
	ID string `json:"id"`
}

type EmailOnlyRequest struct {
	Email string `json:"email"`
}

type EmailAndAccTypeRequest struct {
	Email       string `json:"email"`
	AccountType int    `json:"account_type"`
}

type FileUploadRequest struct {
	File multipart.File `json:"file,omitempty"`
}

type CSVFileUploadRequest struct {
	CSVFile *multipart.FileHeader `form:"file"`
}

package rq

import "mime/multipart"

type IDOnlyRequest struct {
	ID string `json:"id"`
}

type IDsRequest struct {
	IDs []string `json:"ids"`
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

type ImageUploadRequest struct {
	ImageFile *multipart.FileHeader `form:"file"`
}

type CSVFileUploadRequest struct {
	CSVFile *multipart.FileHeader `form:"file"`
}

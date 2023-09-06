package rq

type SchoolRequest struct {
	ID   *string `json:"id,omitempty"`
	Name string  `json:"name"`
	Code string  `json:"code"`
}

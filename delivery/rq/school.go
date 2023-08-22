package rq

type SchoolRequest struct {
	ID   *string `json:"id,omitempty"`
	Name string  `json:"name"`
}

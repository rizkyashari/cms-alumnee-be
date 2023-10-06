package rs

type PaginationResponse[K, V interface{}] struct {
	MaxPage         int `json:"max_page"`
	CurrentPage     int `json:"current_page"`
	RowCount        int `json:"row_count"`
	AvailableFilter K   `json:"available_filter"`

	Data []V `json:"data"`
}

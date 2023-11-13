package rs

type MidtransCredentialsResponse struct {
	ServerKey      string `json:"server_key"`
	Environment    int    `json:"environment"`
	TransactionAPI string `json:"transaction_api"`
	ClientKey      string `json:"client_key"`
	SnapJSUrl      string `json:"snapjs_url"`
	SchoolID       string `json:"school_id"`
}

type MidtransFrontendResponse struct {
	ClientKey string `json:"client_key"`
	SnapJSUrl string `json:"snapjs_url"`
}

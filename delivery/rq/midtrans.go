package rq

type MidtransCredentials struct {
	ServerKey      string `json:"server_key"`
	ClientKey      string `json:"client_key"`
	Environment    int    `json:"environment"`
	TransactionAPI string `json:"transaction_api"`
	SnapJSUrl      string `json:"snapjs_url"`
}

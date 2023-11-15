package model

type ZabbixResponse struct {
	Jsonrpc   string      `json:"jsonrpc"`
	Result    interface{} `json:"result"`
	RawResult []byte
	Error     Error `json:"error"`
	Id        int   `json:"id"`
}

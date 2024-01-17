package model

type ZabbixRequest struct {
	Jsonrpc string       `json:"jsonrpc"`
	Method  string       `json:"method"`
	Params  ZabbixParams `json:"params"`
	Auth    string       `json:"auth,omitempty"`
	Id      int          `json:"id"`
}

type ZabbixParams map[string]interface{}

type ZabbixArrayRequest struct {
	Jsonrpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Auth    string   `json:"auth,omitempty"`
	Id      int      `json:"id"`
}

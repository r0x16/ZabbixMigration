package model

type ZabbixRequest struct {
	Jsonrpc string       `json:"jsonrpc"`
	Method  string       `json:"method"`
	Params  ZabbixParams `json:"params"`
	Auth    string       `json:"auth,omitempty"`
	Id      int          `json:"id"`
}

type ZabbixParams map[string]interface{}

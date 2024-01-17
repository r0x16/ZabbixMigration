package domain

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"

type ZabbixConnectorProvider interface {
	Connect(user string, password string) *model.Error
	Request(body *model.ZabbixRequest) (*model.ZabbixResponse, *model.Error)
	ArrayRequest(body *model.ZabbixArrayRequest) (*model.ZabbixResponse, *model.Error)
	UnauthorizedBody(method string, params model.ZabbixParams) *model.ZabbixRequest
	Body(method string, params model.ZabbixParams) *model.ZabbixRequest
	ArrayBody(method string, params []string) *model.ZabbixArrayRequest
	GetVersion() (model.ZabbixVersion, *model.Error)
}

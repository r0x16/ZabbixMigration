package zabbix

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
)

type ZabbixConnector64 struct {
	*ZabbixConnector
}

var _ domain.ZabbixConnectorProvider = &ZabbixConnector64{}

const (
	VERSION_64 model.ZabbixVersion = 6
	VERSION_60 model.ZabbixVersion = VERSION_64
)

/**
 * API is a function that returns a ZabbixConnector
 * It receives a string with the URL of the Zabbix API
 * It returns a pointer to a ZabbixConnector
 */
func API64(url string) *ZabbixConnector64 {
	return &ZabbixConnector64{
		ZabbixConnector: API(url),
	}
}

/**
 * Connect is a function that connects to the Zabbix API version 6.4
 * Sets the token in the ZabbixConnector for future requests
 */
func (z *ZabbixConnector64) Connect(user string, password string) *model.Error {
	return z.baseConnect(model.ZabbixParams{
		"username": user,
		"password": password,
	})
}

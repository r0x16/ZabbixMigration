package zabbix

import (
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/domain/model"
	"git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/domain"
)

type ZabbixConnector40 struct {
	*ZabbixConnector
}

var _ domain.ZabbixConnectorProvider = &ZabbixConnector40{}

const (
	VERSION_40 model.ZabbixVersion = 4
)

/**
 * API is a function that returns a ZabbixConnector
 * It receives a string with the URL of the Zabbix API
 * It returns a pointer to a ZabbixConnector
 */
func API40(url string) *ZabbixConnector40 {
	return &ZabbixConnector40{
		ZabbixConnector: API(url),
	}
}

/**
 * Connect is a function that connects to the Zabbix API version 4.0
 * Sets the token in the ZabbixConnector for future requests
 */
func (z *ZabbixConnector40) Connect(user string, password string) *model.Error {
	return z.baseConnect(model.ZabbixParams{
		"user":     user,
		"password": password,
	})
}

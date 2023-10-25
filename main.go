package main

import "git.tnschile.com/sistemas/zabbix/zabbix-migration/src/shared/infraestructure"

func main() {
	application := &infraestructure.Main{}
	application.RunServices()
}

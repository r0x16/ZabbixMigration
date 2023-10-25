package main

import "git.tnschile.com/sistemas/tnsgo/raidark/src/shared/infraestructure"

func main() {
	application := &infraestructure.Main{}
	application.RunServices()
}

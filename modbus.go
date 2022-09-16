package main

import (
	"fmt"
	"github.com/goburrow/modbus"
	"log"
)

func mymodbustest() {
	log.Println("Started")

	cli := modbus.TCPClient("heatpump.ezplanet.org:502")
	res, err := cli.ReadHoldingRegisters(400, 2)
	fmt.Println(res)
	fmt.Println(err)
}

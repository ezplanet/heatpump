package base

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const (
	vitocalModbusAddrKey     string = "VITOCAL_MODBUS_ADDR"
	vitocalModbusAddrDefault int    = 1

	vitocalModbusTcpKey     string = "VITOCAL_MODBUS_TCP"
	vitocalModbusTcpDefault string = "heatpump:502"
)

var (
	VitocalModbusAddr int
	VitocalModbusTcp  string
)

func init() {
	var err error

	err = godotenv.Load() //Load .env file
	if err != nil {
		log.Print(err)
	}

	if len(os.Getenv(vitocalModbusAddrKey)) == 0 {
		VitocalModbusAddr = vitocalModbusAddrDefault
	} else {
		VitocalModbusAddr, err = strconv.Atoi(os.Getenv(vitocalModbusAddrKey))
		if err != nil {
			VitocalModbusAddr = vitocalModbusAddrDefault
		}
	}

	VitocalModbusTcp = os.Getenv(vitocalModbusTcpKey)
	if len(VitocalModbusTcp) <= 0 {
		VitocalModbusTcp = vitocalModbusTcpDefault
	}
}

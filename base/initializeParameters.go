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

	baseSHMKey     string = "BASE_SHM"
	baseSHMDefault string = "/dev/shm"

	rawLogKey     string = "RAWLOG"
	rawLogDefault bool   = false
)

var (
	VitocalModbusAddr int
	VitocalModbusTcp  string
	BaseSHM           string
	RawLog            bool
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

	BaseSHM = os.Getenv(baseSHMKey)
	if len(BaseSHM) <= 0 {
		BaseSHM = baseSHMDefault
	} else {
		if BaseSHM[len(BaseSHM)-1:] != "/" {
			BaseSHM = BaseSHM + "/"
		}
	}

	if len(os.Getenv(rawLogKey)) == 0 {
		RawLog = rawLogDefault
	} else {
		RawLog, err = strconv.ParseBool(os.Getenv(rawLogKey))
		if err != nil {
			RawLog = rawLogDefault
		}
	}
	log.Print("RAWLOG: ", RawLog)
}

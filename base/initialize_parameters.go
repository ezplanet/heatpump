package base

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

const (
	mqttServerKey     string = "MQTT_SERVER"
	mqttServerDefault string = "ssl://lambo.ezplanet.org:8883"

	mqttClientIdKey     string = "MQTT_CLIENT_ID"
	mqttClientIdDefault string = "heatpump"

	mqttTopicKey     string = "MQTT_TOPIC"
	mqttTopicDefault string = "climatico/vitocal"

	vitocalModbusAddrKey     string = "MODBUS_ADDR"
	vitocalModbusAddrDefault int    = 1

	vitocalModbusTcpKey     string = "MODBUS_TCP"
	vitocalModbusTcpDefault string = "heatpump:502"

	modbusConnectionTimeoutMinutesKey     string = "MODBUS_CONNECTION_TIMEOUT_MINUTES"
	modbusConnectionTimeoutMinutesDefault int    = 60

	standbyThrottleSecondsKey     string = "STANDBY_THROTTLE_SECONDS"
	standbyThrottleSecondsDefault int    = 60

	baseSHMKey     string = "BASE_SHM"
	baseSHMDefault string = "/dev/shm"

	rawLogKey     string = "RAWLOG"
	rawLogDefault bool   = false
)

var (
	MqttServer                     string
	MqttClientId                   string
	MqttTopic                      string
	VitocalModbusAddr              int
	VitocalModbusTcp               string
	ModbusConnectionTimeoutMinutes int
	StandbyThrottleSeconds         float64
	BaseSHM                        string
	RawLog                         bool
)

func init() {
	var err error

	err = godotenv.Load() //Load .env file
	if err != nil {
		log.Print(err)
	}

	MqttServer = os.Getenv(mqttServerKey)
	if len(MqttServer) <= 0 {
		MqttServer = mqttServerDefault
	}

	MqttClientId = os.Getenv(mqttClientIdKey)
	if len(MqttServer) <= 0 {
		MqttClientId = mqttClientIdDefault
	}

	MqttTopic = os.Getenv(mqttTopicKey)
	if len(MqttServer) <= 0 {
		MqttTopic = mqttTopicDefault
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

	if len(os.Getenv(modbusConnectionTimeoutMinutesKey)) == 0 {
		ModbusConnectionTimeoutMinutes = modbusConnectionTimeoutMinutesDefault
	} else {
		ModbusConnectionTimeoutMinutes, err = strconv.Atoi(os.Getenv(modbusConnectionTimeoutMinutesKey))
		if err != nil {
			ModbusConnectionTimeoutMinutes = modbusConnectionTimeoutMinutesDefault
		}
	}

	if len(os.Getenv(standbyThrottleSecondsKey)) == 0 {
		StandbyThrottleSeconds = float64(standbyThrottleSecondsDefault)
	} else {
		StandbyThrottleSecondsInt, err := strconv.Atoi(os.Getenv(standbyThrottleSecondsKey))
		if err != nil {
			StandbyThrottleSeconds = float64(standbyThrottleSecondsDefault)
		} else {
			StandbyThrottleSeconds = float64(StandbyThrottleSecondsInt)
		}
	}

	BaseSHM = os.Getenv(baseSHMKey)
	if len(BaseSHM) <= 0 {
		BaseSHM = baseSHMDefault
	} else {
		err := os.MkdirAll(BaseSHM, 0755)
		if err != nil {
			BaseSHM = baseSHMDefault
		}
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

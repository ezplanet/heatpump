package decoder

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"heatpump/base"
	"heatpump/domain"
	"heatpump/mqtt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

const (
	MODBUS_READ  uint8 = 0x03
	ERRORS       uint8 = 0x08
	TEMPERATURES uint8 = 0x04
	STATES       uint8 = 0x02
	MACHINE      uint8 = 0x01
	COMPLETE     uint8 = 0x0F

	OFF byte = 0x00
	ON  byte = 0x01

	//MACHINE
	// byte 1
	COMPRESSOR_OIL_HEATER byte = 0x80
	// byte 2
	COMPRESSOR_STARTING byte = 0x01
	COMPRESSOR_RUNNING  byte = 0x04
	COMPRESSOR_THRUST   byte = 0x08
	CIRCULATION_PUMP_ON byte = 0x40
	// byte 7 and 8
	COMPRESSOR_ACTIVE       uint16 = 0x8000
	CIRCULATION_PUMP_ACTIVE uint16 = 0x0601

	//STATUS
	// byte 3
	STATUS_STANDBY             byte = 0x02
	STATUS_COMPRESSOR_REQUIRED byte = 0x10
	STATUS_DEFROST_STARTING    byte = 0x30
	STATUS_DEFROST_ACTIVE      byte = 0x50

	// byte 4
	VITOCAL_OFF       byte = 0x00
	VITOCAL_AUTO_COOL byte = 0x01
	VITOCAL_AUTO_HEAT byte = 0x02

	// byte 7
	HEAT        byte = 0x40
	COOL        byte = 0x80
	COOL_MANUAL byte = 0xc0

	// BASE
	VITOCAL_POWERED       string = "VitocalPowered"
	VITOCAL_PUMP_ON       string = "VitocalPumpOn"
	VITOCAL_STATUS_ON     string = "VitocalStatusOn"
	VITOCAL_COMPRESSOR_ON string = "VitocalCompressorOn"
	VITOCAL_MODE_COOL     string = "VitocalModeCool"
)

var (
	vitocalPowered    uint8 = 0xFF
	vitocalPump       uint8 = 0xFF
	vitocalStatus     uint8 = 0xFF
	vitocalCompressor uint8 = 0xFF
	vitocalModeCool   uint8 = 0xFF
)

func Decode(c net.Conn) error {
	defer c.Close()

	var lastTime time.Time
	buf := make([]byte, 256) // using small buffer
	var template uint8 = 0
	var temperatures string
	var states string
	var machine string
	var errors string
	var raw_temperatures string
	var vitocal domain.Vitocal

	for {
		c.SetReadDeadline(time.Now().Add(15 * time.Second))
		size, err := c.Read(buf)
		if err != nil {
			// If we have a connection, but there is no data stream then we assume that the heatpump is not powered
			if os.IsTimeout(err) {
				if vitocalPowered >= 1 {
					cmd := exec.Command("/bin/rm", "-f", base.BaseSHM+VITOCAL_POWERED,
						base.BaseSHM+VITOCAL_STATUS_ON, base.BaseSHM+VITOCAL_PUMP_ON, base.BaseSHM+VITOCAL_COMPRESSOR_ON,
						base.BaseSHM+VITOCAL_MODE_COOL)
					err := cmd.Run()
					if err != nil {
						log.Printf("error removing vitocal state files: %s\n", err)
					} else {
						vitocalPowered = OFF
						vitocalStatus = OFF
						vitocalCompressor = OFF
						vitocalPump = OFF
						vitocalModeCool = OFF
					}
				}
				continue
			}
			if err != io.EOF {
				log.Println("error reading MODBUS stream", err)
				return err
			}
		}

		// We can read the data stream therefore the heatpump is powered
		vitocalPowered = setVitocalStateOn(vitocalPowered, VITOCAL_POWERED)

		// Filter by known responses and CRC CHECK
		// If the third byte (buf[2]) is equal record length less 5 then this is likely a response
		if size > 2 && int(buf[2]) == (size-5) && int(buf[0]) == base.VitocalModbusAddr && uint8(buf[1]) == MODBUS_READ {
			checksum := crc16(buf, size)
			if checksum[0] != buf[size-2] || checksum[1] != buf[size-1] {
				fmt.Println("CRC ERROR")
				fmt.Println(size, buf[2], checksum, buf[size-2], buf[size-1], buf)
				continue
			}
			//fmt.Println(size, buf[2], checksum, buf[size-2], buf[size-1], buf)

			// TEMPERATURES and PRESSURES - Address 0x018f
			if size == 105 && buf[2] == 100 && (template&TEMPERATURES) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)
				temperatureIn := float32(value[1]) / 10
				temperatureOut := float32(value[2]) / 10
				temperatureExt := float32(value[29]) / 10
				ingressoComp := float32(value[23]) / 10
				scaricoComp := float32(value[34]) / 10
				suctionPressure := float32(value[15]) / 100
				condensationPressure := float32(value[7]) / 100
				vitocal.Temperatures.WaterIn = fmt.Sprintf("%.1f", float32(value[1])/10)
				vitocal.Temperatures.WaterOut = fmt.Sprintf("%.1f", float32(value[2])/10)
				vitocal.Temperatures.External = fmt.Sprintf("%.1f", float32(value[29])/10)
				vitocal.Temperatures.CompressorIn = fmt.Sprintf("%.1f", float32(value[23])/10)
				vitocal.Temperatures.CompressorOut = fmt.Sprintf("%.1f", float32(value[34])/10)
				vitocal.PressureCondensation = int(value[7])
				vitocal.PressureSuction = int(value[15])
				temperatures = fmt.Sprintf("Temp: wtr_in=%.1f wtr_out=%.1f ext=%.1f cmp_in=%.1f cmp_out=%.1f - Press: suct=%.2f cond=%.2f",
					temperatureIn, temperatureOut, temperatureExt, ingressoComp, scaricoComp,
					suctionPressure, condensationPressure)

				if base.RawLog {
					raw_temperatures = ""
					for i := 0; i < len(value)-2; i++ {
						raw_temperatures = fmt.Sprintf("%s%04x ", raw_temperatures, value[i])
					}
				}
				template |= TEMPERATURES
			}

			// STATES - Address 0x1c2e - Size 11
			if size == 27 && buf[2] == 22 && (template&STATES) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)
				// if bit 2 = 0 standby otherwise on
				if buf[3]&STATUS_STANDBY == 0 {
					vitocal.Status = domain.ON
					vitocalStatus = setVitocalStateOn(vitocalStatus, VITOCAL_STATUS_ON)
				} else {
					vitocal.Status = domain.OFF
					vitocalStatus = setVitocalStateOff(vitocalStatus, VITOCAL_STATUS_ON)
				}
				if buf[3]&STATUS_COMPRESSOR_REQUIRED == STATUS_COMPRESSOR_REQUIRED {
					vitocal.CompressorRequired = true
					vitocalCompressor = setVitocalStateOn(vitocalCompressor, VITOCAL_COMPRESSOR_ON)
				} else {
					vitocal.CompressorRequired = false
					vitocalCompressor = setVitocalStateOff(vitocalCompressor, VITOCAL_COMPRESSOR_ON)
				}

				// DEFROST
				if buf[3]&STATUS_DEFROST_STARTING == STATUS_DEFROST_STARTING {
					vitocal.Defrost = domain.DEFROST_STARTING
				} else if buf[3]&STATUS_DEFROST_ACTIVE == STATUS_DEFROST_ACTIVE {
					vitocal.Defrost = domain.DEFROST_ACTIVE
				} else {
					vitocal.Defrost = domain.DEFROST_INACTIVE
				}
				// ToDo CONTROL_MODE_HEAT and CONTROL_MODE_COOL have same value, need to manage difference in app
				switch buf[4] {
				case VITOCAL_OFF:
					vitocal.ControlMode = domain.CONTROL_MODE_OFF
				case VITOCAL_AUTO_COOL:
					vitocal.ControlMode = domain.CONTROL_MODE_COOL
				case VITOCAL_AUTO_HEAT:
					vitocal.ControlMode = domain.CONTROL_MODE_HEAT
				}
				switch buf[7] {
				case HEAT:
					vitocal.Mode = domain.MODE_HEAT
					vitocalModeCool = setVitocalStateOff(vitocalModeCool, VITOCAL_MODE_COOL)
				case COOL:
					vitocal.Mode = domain.MODE_COOL
					vitocalModeCool = setVitocalStateOn(vitocalModeCool, VITOCAL_MODE_COOL)
				}
				vitocal.CompressorHz = int(value[5])
				vitocal.FanSpeed = int(value[6])
				vitocal.PumpSpeed = int(value[7])
				vitocal.Hours = int(value[8])
				states = fmt.Sprintf("%04x %04x %04x", value[0], value[1], value[2])
				for i := 3; i < len(value)-2; i++ {
					states = fmt.Sprintf("%s %5d ", states, value[i])
				}
				states = fmt.Sprintf("%s %04x %04x", states, value[9], value[10])
				template |= STATES
			}

			// MACHINE - Address 0x01e0 - Size  3
			if size == 11 && buf[2] == 6 && (template&MACHINE) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)

				if buf[3]&ON == 0 {
					if vitocal.CompressorRequired {
						if buf[4]&COMPRESSOR_STARTING == COMPRESSOR_STARTING {
							vitocal.CompressorStatus = domain.STARTING2
						} else {
							vitocal.CompressorStatus = domain.STARTING
						}
					} else {
						vitocal.CompressorStatus = domain.OFF
					}
				} else {
					if buf[3]&OFF == 0 {
						vitocal.CompressorStatus = domain.ON
					}
				}
				if buf[3]&COMPRESSOR_OIL_HEATER == COMPRESSOR_OIL_HEATER {
					vitocal.OilHeater = domain.ON
				} else {
					vitocal.OilHeater = domain.OFF
				}
				if buf[4]&COMPRESSOR_THRUST == COMPRESSOR_THRUST {
					vitocal.CompressorThrust = domain.ON
				} else {
					vitocal.CompressorThrust = domain.OFF
				}
				if value[2]&CIRCULATION_PUMP_ACTIVE == CIRCULATION_PUMP_ACTIVE {
					vitocal.PumpStatus = domain.ON
					vitocalPump = setVitocalStateOn(vitocalPump, VITOCAL_PUMP_ON)
				} else {
					vitocal.PumpStatus = domain.OFF
					vitocalPump = setVitocalStateOff(vitocalPump, VITOCAL_PUMP_ON)
				}
				for i := 0; i < len(value); i++ {
					machine = fmt.Sprintf("%s %04x", machine, value[i])
				}
				template |= MACHINE
			}

			// ERRORS - Address 0x03ca - Size  5
			if size == 15 && buf[2] == 10 && (template&ERRORS) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)
				for i := 0; i < len(value); i++ {
					errors = fmt.Sprintf("%s %d", errors, value[i])
				}
				vitocal.Errors.Error1 = value[0]
				vitocal.Errors.Error2 = value[1]
				vitocal.Errors.Error3 = value[2]
				vitocal.Errors.Error4 = value[3]
				vitocal.Errors.Error5 = value[4]
				template |= ERRORS
			}
		}

		vitocal.Timestamp = time.Now()

		buf = make([]byte, 256)
		if (template & COMPLETE) == COMPLETE {
			//prettyJSON, err := json.MarshalIndent(vitocal, "", "   ")
			linearJSON, err := json.Marshal(vitocal)
			if err != nil {
				log.Fatal("failed to generate JSON")
			} else {
				// Throttle down to 1 message every StandbyThrottleSeconds when the heat pump is in STAND BY
				if vitocal.Status == domain.ON || vitocal.PumpStatus == domain.ON ||
					vitocal.Timestamp.Sub(lastTime).Seconds() > base.StandbyThrottleSeconds {
					log.Printf("%s - %s - %s -%s\n", machine, states, temperatures, errors)
					if base.RawLog {
						fmt.Printf("%s  %s\n", vitocal.Timestamp.Format("2006/01/02 15:04:05"), raw_temperatures)
						//fmt.Printf("%s\n", string(prettyJSON))
					}
					err := mqtt.Publish(base.MqttTopic, true, string(linearJSON))
					if err != nil {
						log.Print("MQTT publish Error: ", err)
					}
					lastTime = vitocal.Timestamp
				} else {
					//fmt.Printf("waiting %s\n", vitocal.Timestamp.Sub(lastTime))
				}
			}
			template = 0
			machine = ""
			states = ""
			temperatures = ""
			errors = ""
		}
	}
	return fmt.Errorf("modbus data stream reading interrupted")
}

func setVitocalStateOn(state uint8, file string) uint8 {
	if state == OFF || state == 0xFF {
		_, err := os.Create(base.BaseSHM + file)
		if err != nil {
			fmt.Println("Error creating file: ", base.BaseSHM+file)
			return state
		} else {
			return ON
		}
	} else {
		return state
	}
}

func setVitocalStateOff(state uint8, file string) uint8 {
	if state == ON || state == 0xFF {
		cmd := exec.Command("/bin/rm", "-f", base.BaseSHM+file)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Error removing vitocal state file %s: %s\n", file, err)
			return state
		} else {
			return OFF
		}
	} else {
		return state
	}
}

func getValues(buf []byte, size int) []uint16 {
	var values []uint16
	offset := 3
	for i := 0; i < size/2; i++ {
		address := offset + i*2
		byte16 := []byte{buf[address], buf[address+1]}
		values = append(values, binary.BigEndian.Uint16(byte16))
	}
	return values
}

func crc16(buf []byte, size int) [2]byte {
	//fmt.Printf("Checksum: %01x %01x - %d\n", uint8(buf[size+4]), buf[size+3], size)
	var checksum [2]byte
	var regCRC uint16 = 0xFFFF
	for i := 0; i < size-2; i++ {
		regCRC ^= uint16(buf[i])
		for j := 0; j < 8; j++ {
			if (regCRC & 0x01) == 1 {
				regCRC = (regCRC >> 1) ^ 0xA001
			} else {
				regCRC = regCRC >> 1
			}
		}
	}
	checksum[1] = byte((regCRC >> 8) & 0xFF)
	checksum[0] = byte(regCRC & 0xFF)
	return checksum
}

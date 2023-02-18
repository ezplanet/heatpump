package main

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
	MODBUS_READ uint8 = 0x03
	//UNIDENTIFIED uint8 = 0x10
	ERRORS       uint8 = 0x08
	TEMPERATURES uint8 = 0x04
	STATES       uint8 = 0x02
	MACHINE      uint8 = 0x01
	COMPLETE     uint8 = 0x0F
	//COMPLETE     uint8 = 0x1F

	OFF byte = 0x00
	ON  byte = 0x01

	//MACHINE
	// byte 1
	COMPRESSOR_OIL_HEATER = 0x90
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
	BASE_SHM              string = "/dev/shm/"
	VITOCAL_POWERED       string = "VitocalPowered"
	VITOCAL_PUMP_ON       string = "VitocalPumpOn"
	VITOCAL_STATUS_ON     string = "VitocalStatusOn"
	VITOCAL_COMPRESSOR_ON string = "VitocalCompressorOn"
)

var (
	vitocalPowered    uint8 = 0xFF
	vitocalPump       uint8 = 0xFF
	vitocalStatus     uint8 = 0xFF
	vitocalCompressor uint8 = 0xFF
)

func main() {

	var slaveAddress byte = 1
	var functionCode byte = 3
	var startAdress uint16 = 0
	var numberOfPoints uint16 = 2
	var lastTime time.Time

	// build Modbus RTU message
	var frame [8]byte
	frame[0] = slaveAddress
	frame[1] = functionCode
	frame[2] = byte(startAdress >> 8)
	frame[3] = byte(startAdress)
	frame[4] = byte(numberOfPoints >> 8)
	frame[5] = byte(numberOfPoints)
	checksum := crc16(frame[:], 8)
	frame[6] = checksum[0]
	frame[7] = checksum[1]
	fmt.Print("Checksum test: ")
	for i := 0; i < 8; i++ {
		fmt.Printf("%02x ", frame[i])
	}
	fmt.Println()

	c, err := net.Dial("tcp", base.VitocalModbusTcp)
	if err != nil {
		fmt.Printf("Error %s trying to connect to %s\n", err, base.VitocalModbusTcp)
		return
	}
	defer c.Close()

	buf := make([]byte, 256) // using small buffer
	var template uint8 = 0
	var temperatures string
	var states string
	var machine string
	var errors string
	//var unidentified string
	//var temp_unidentified string
	var vitocal domain.Vitocal
	var errorCount int = 0

	for {
		c.SetReadDeadline(time.Now().Add(15 * time.Second))
		size, err := c.Read(buf)
		if err != nil {
			// If we cannot read the data stream then we assume that the heatpump is not powered
			if os.IsTimeout(err) {
				//fmt.Printf("Timeout: %s\n", err)
				if vitocalPowered >= 1 {
					cmd := exec.Command("/bin/rm", "-f", BASE_SHM+VITOCAL_POWERED,
						BASE_SHM+VITOCAL_STATUS_ON, BASE_SHM+VITOCAL_PUMP_ON, BASE_SHM+VITOCAL_COMPRESSOR_ON)
					err := cmd.Run()
					if err != nil {
						fmt.Printf("Error removing vitocal state files: %s\n", err)
					} else {
						vitocalPowered = OFF
						vitocalStatus = OFF
						vitocalCompressor = OFF
						vitocalPump = OFF
					}
				}
				continue
			}
			if err != io.EOF {
				fmt.Println("Error reading MODBUS stream", err)
				time.Sleep(60 * time.Second)
				c, err = net.Dial("tcp", base.VitocalModbusTcp)
				if err != nil {
					errorCount++
					fmt.Printf("Error %s trying to connect to %s\n", err, base.VitocalModbusTcp)
				}
				if errorCount > 60 {
					break
				} else {
					continue
				}
			}
			break
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

			// TEMPERATURES - Address 0x018f
			if size == 105 && buf[2] == 100 && (template&TEMPERATURES) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)
				temperatureIn := float32(value[1]) / 10
				temperatureOut := float32(value[2]) / 10
				temperatureExt := float32(value[29]) / 10
				scaricoComp := float32(value[34]) / 10
				vitocal.Temperatures.Input = fmt.Sprintf("%.1f", float32(value[1])/10)
				vitocal.Temperatures.Output = fmt.Sprintf("%.1f", float32(value[2])/10)
				vitocal.Temperatures.External = fmt.Sprintf("%.1f", float32(value[29])/10)
				vitocal.Temperatures.Compressor = fmt.Sprintf("%.1f", float32(value[34])/10)
				vitocal.PressureHigh = int(value[7])
				vitocal.PressureLow = int(value[15])
				temperatures = fmt.Sprintf("Temp: in=%.1f out=%.1f ext=%.1f comp=%.1f - Press: high=%d low=%d",
					temperatureIn, temperatureOut, temperatureExt, scaricoComp,
					value[7], value[15])

				//temp_unidentified = ""
				//for i := 0; i < len(value)-2; i++ {
				//	temp_unidentified = fmt.Sprintf("%s %04x ", temp_unidentified, value[i])
				//}
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
				case COOL:
					vitocal.Mode = domain.MODE_COOL
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
					vitocal.CompressorThrust = true
				} else {
					vitocal.CompressorThrust = false
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

			// Address 0xc288 - Size 1
			//if size == 7 && buf[2] == 2 && (template&UNIDENTIFIED) == 0 {
			//	unidentified = fmt.Sprintf("%02x %02x", buf[3], buf[4])
			//	template |= UNIDENTIFIED
			//}
		}

		vitocal.Timestamp = time.Now()

		buf = make([]byte, 256)
		if (template & COMPLETE) == COMPLETE {
			//prettyJSON, err := json.MarshalIndent(vitocal, "", "   ")
			linearJSON, err := json.Marshal(vitocal)
			if err != nil {
				log.Fatal("Failed to generate JSON")
			} else {
				// Throttle down to 1 message every minute when the heat pump is in STAND BY
				if vitocal.Status == domain.ON || vitocal.PumpStatus == domain.ON || vitocal.Timestamp.Sub(lastTime).Seconds() > 60 {
					log.Printf("%s - %s - %s -%s\n", machine, states, temperatures, errors)
					//fmt.Printf("%s  %s - %s\n", vitocal.Timestamp.Format("2006/02/01 15:04:05"), unidentified, temp_unidentified)
					//fmt.Printf("%s\n", string(prettyJSON))
					mqtt.Publish(mqtt.VitocalTopic, true, string(linearJSON))
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
}

func setVitocalStateOn(state uint8, file string) uint8 {
	if state == OFF || state == 0xFF {
		_, err := os.Create(BASE_SHM + file)
		if err != nil {
			fmt.Println("Error creating file: ", BASE_SHM+file)
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
		cmd := exec.Command("/bin/rm", "-f", BASE_SHM+file)
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

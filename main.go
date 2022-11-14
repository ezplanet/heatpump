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
	"time"
)

const (
	MODBUS_READ  uint8 = 3
	ERRORS       uint8 = 8
	TEMPERATURES uint8 = 4
	STATES       uint8 = 2
	MACHINE      uint8 = 1
	COMPLETE     uint8 = 15

	OFF     byte = 0x00
	ON      byte = 0x01
	STANDBY byte = 0x02
	COMPREQ byte = 0x10

	//STATUS byte 4
	VITOCAL_OFF  byte = 0x00
	VITOCAL_ON   byte = 0x01
	VITOCAL_AUTO byte = 0x02

	// STATUS byte 7
	HEAT        byte = 0x40
	COOL        byte = 0x80
	COOL_MANUAL byte = 0xc0

	// STATUS json

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
	var vitocal domain.Vitocal

	for {
		size, err := c.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("read error", err)
			}
			break
		}

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
				template |= TEMPERATURES
			}
			if size == 27 && buf[2] == 22 && (template&STATES) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)
				// if bit 2 = 0 standby otherwise on
				if buf[3]&STANDBY == 0 {
					vitocal.Status = domain.ON
				} else {
					vitocal.Status = domain.OFF
				}
				if buf[3]&COMPREQ == COMPREQ {
					vitocal.CompressorRequired = true
				} else {
					vitocal.CompressorRequired = false
				}
				switch buf[4] {
				case VITOCAL_OFF:
					vitocal.ControlMode = domain.CONTROL_MODE_OFF
				case VITOCAL_ON:
					vitocal.ControlMode = domain.CONTROL_MODE_ON
				case VITOCAL_AUTO:
					vitocal.ControlMode = domain.CONTROL_MODE_AUTO
				}
				switch buf[7] {
				case HEAT:
					vitocal.Mode = domain.MODE_HEAT
				case COOL:
					vitocal.Mode = domain.MODE_COOL
				}
				vitocal.FanSpeed = int(value[5])
				vitocal.CompressorHz = int(value[6])
				vitocal.PumpSpeed = int(value[7])
				vitocal.Hours = int(value[8])
				states = fmt.Sprintf("%04x %04x %04x", value[0], value[1], value[2])
				for i := 3; i < len(value)-2; i++ {
					states = fmt.Sprintf("%s %5d ", states, value[i])
				}
				states = fmt.Sprintf("%s %04x %04x", states, value[9], value[10])
				template |= STATES
			}
			if size == 11 && buf[2] == 6 && (template&MACHINE) == 0 {
				dataSize := int(buf[2])
				value := getValues(buf, dataSize)

				if buf[3]&ON == 0 {
					if vitocal.CompressorRequired {
						vitocal.CompressorStatus = domain.STARTING
					} else {
						vitocal.CompressorStatus = domain.OFF
					}
				} else {
					if buf[3]&OFF == 0 {
						vitocal.CompressorStatus = domain.ON
					}
				}
				if value[2] == 0x601 || value[2] == 0x8601 {
					vitocal.PumpStatus = domain.ON
				} else {
					vitocal.PumpStatus = domain.OFF
				}
				for i := 0; i < len(value); i++ {
					machine = fmt.Sprintf("%s %04x", machine, value[i])
				}
				template |= MACHINE
			}
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
				log.Fatal("Failed to generate JSON")
			} else {
				// Throttle down to 1 message every minute when the heat pump is in STAND BY
				if vitocal.Status == domain.ON || vitocal.PumpStatus == domain.ON || vitocal.Timestamp.Sub(lastTime).Seconds() > 60 {
					log.Printf("%s - %s - %s -%s\n", machine, states, temperatures, errors)
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

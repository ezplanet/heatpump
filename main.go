package main

import (
	"heatpump/base"
	"heatpump/decoder"
	"io"
	"log"
	"net"
	"time"
)

// Connects to the heatpump modbus service and then hands the connection to the decoder
// Retries the connection in case of error up to the defined timeout
func main() {
	var errorCount int = 0
	for {
		conn, err := net.Dial("tcp", base.VitocalModbusTcp)
		if err != nil {
			errorCount++
			log.Printf("error: '%s' trying to connect to: '%s'\n", err, base.VitocalModbusTcp)
			if errorCount > base.ModbusConnectionTimeoutMinutes {
				log.Fatalf("failed to connect to %s for %d minutes\n", base.VitocalModbusTcp,
					base.ModbusConnectionTimeoutMinutes)
				break
			} else {
				time.Sleep(60 * time.Second)
				continue
			}
		}
		errorCount = 0
		err = decoder.Decode(conn)
		if err == io.EOF {
			log.Fatalf("end of data from %s", base.VitocalModbusTcp)
			break
		} else {
			log.Println("error:", err)
		}
		conn.Close()
	}
}

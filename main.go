// Copyright 2023 by mauro@ezplanet.org (Mauro Mozzarelli)
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"io"
	"net"
	"time"

	"heatpump/base"
	"heatpump/decoder"
	"heatpump/logger"
)

// Connects to the heatpump modbus service and then hands the connection to the decoder
// Retries the connection in case of error up to the defined timeout
func main() {
	err := logger.Init(base.LogLevel, base.LogFile)
	if err != nil {
		panic(err)
	}
	var errorCount int = 0
	for {
		conn, err := net.Dial("tcp", base.VitocalModbusTcp)
		if err != nil {
			errorCount++
			logger.Log.Error().Msgf("error: '%s' trying to connect to: '%s'\n", err.Error(), base.VitocalModbusTcp)
			if errorCount > base.ModbusConnectionTimeoutMinutes {
				logger.Log.Fatal().Msgf("failed to connect to %s for %d minutes\n", base.VitocalModbusTcp,
					base.ModbusConnectionTimeoutMinutes)
				break
			} else {
				time.Sleep(60 * time.Second)
				continue
			}
		}
		errorCount = 0
		logger.Log.Trace().Msg("now decoding data stream...")
		err = decoder.Decode(conn)
		if err == io.EOF {
			logger.Log.Fatal().Msgf("end of data from %s", base.VitocalModbusTcp)
			break
		} else {
			logger.Log.Error().Msgf("error: %s", err.Error())
		}
		conn.Close()
	}
}

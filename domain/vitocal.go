//BSD 2-Clause License
//
//Copyright (c) 2023, Mauro Mozzarelli
//
//Redistribution and use in source and binary forms, with or without
//modification, are permitted provided that the following conditions are met:
//
//1. Redistributions of source code must retain the above copyright notice, this
//list of conditions and the following disclaimer.
//
//2. Redistributions in binary form must reproduce the above copyright notice,
//this list of conditions and the following disclaimer in the documentation
//and/or other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
//AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
//IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
//FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
//DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
//SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
//CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
//OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
//OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package domain

import (
	"heatpump/domain/vitocal"
	"time"
)

const (
	CONTROL_MODE_OFF  int = 0
	CONTROL_MODE_COOL int = 2
	CONTROL_MODE_HEAT int = 2

	OFF              int = 0
	ON               int = 1
	STARTING         int = 2
	STARTING2        int = 3
	MODE_HEAT        int = 1
	MODE_COOL        int = 2
	DEFROST_INACTIVE int = 0
	DEFROST_STARTING int = 1
	DEFROST_ACTIVE   int = 2
)

type Vitocal struct {
	Timestamp            time.Time            `json:"timestamp"`
	ControlMode          int                  `json:"control_mode"`
	Status               int                  `json:"status"`
	Mode                 int                  `json:"mode"`
	Defrost              int                  `json:"defrost"`
	OilHeater            int                  `json:"oil_heater"`
	CompressorRequired   bool                 `json:"compressor_required"`
	CompressorStatus     int                  `json:"compressor_status"`
	CompressorThrust     int                  `json:"compressor_thrust"`
	CompressorHz         int                  `json:"compressor_hz"`
	PumpStatus           int                  `json:"pump_status"`
	PumpSpeed            int                  `json:"pump_speed"`
	FanSpeed             int                  `json:"fan_speed"`
	Temperatures         vitocal.Temperatures `json:"temperatures"`
	PressureSuction      int                  `json:"pressure_suction"`
	PressureCondensation int                  `json:"pressure_condensation"`
	Hours                int                  `json:"hours"`
	Errors               vitocal.Errors       `json:"errors"`
}

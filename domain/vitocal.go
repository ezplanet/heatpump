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
	MODE_HEAT        int = 1
	MODE_COOL        int = 2
	DEFROST_INACTIVE int = 0
	DEFROST_STARTING int = 1
	DEFROST_ACTIVE   int = 3
)

type Vitocal struct {
	Timestamp          time.Time            `json:"timestamp"`
	ControlMode        int                  `json:"control_mode"`
	Status             int                  `json:"status"`
	Mode               int                  `json:"mode"`
	Defrost            int                  `json:"defrost"`
	OilHeater          int                  `json:"oil_heater"`
	CompressorRequired bool                 `json:"compressor_required"`
	CompressorStatus   int                  `json:"compressor_status"`
	CompressorHz       int                  `json:"compressor_hz"`
	PumpStatus         int                  `json:"pump_status"`
	PumpSpeed          int                  `json:"pump_speed"`
	FanSpeed           int                  `json:"fan_speed"`
	Temperatures       vitocal.Temperatures `json:"temperatures"`
	PressureHigh       int                  `json:"pressure_high"`
	PressureLow        int                  `json:"pressure_low"`
	Hours              int                  `json:"hours"`
	Errors             vitocal.Errors       `json:"errors"`
}

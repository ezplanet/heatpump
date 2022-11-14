package domain

import (
	"heatpump/domain/vitocal"
	"time"
)

const (
	CONTROL_MODE_OFF  int = 0
	CONTROL_MODE_ON   int = 1
	CONTROL_MODE_AUTO int = 2

	OFF       int = 0
	ON        int = 1
	STARTING  int = 2
	MODE_HEAT int = 1
	MODE_COOL int = 2
)

type Vitocal struct {
	Timestamp          time.Time            `json:"timestamp"`
	ControlMode        int                  `json:"control_mode"`
	Status             int                  `json:"status"`
	Mode               int                  `json:"mode"`
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

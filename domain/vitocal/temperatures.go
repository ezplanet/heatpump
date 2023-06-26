package vitocal

type Temperatures struct {
	WaterIn       string `json:"water_in"`
	WaterOut      string `json:"water_out"`
	External      string `json:"external"`
	CompressorIn  string `json:"compressor_in"`
	CompressorOut string `json:"compressor_out"`
}

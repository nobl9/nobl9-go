package v1

type QueryResponse struct {
	TimeSeries []TimeSeries `json:"timeseries"`
}

type TimeSeries struct {
	Measurement string    `json:"measurement"`
	Timestamps  []int64   `json:"timestamps"`
	Values      []float64 `json:"values"`
}

package model

type HistogramMeasurement struct {
	TargetUuid string  `json:"target_uuid"`
	Timestamp  int64   `json:"timestamp"`
	Bucket     float64 `json:"bucket"`
	Count      int64   `json:"count"`
}

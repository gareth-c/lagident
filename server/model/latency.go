package model

type Latency struct {
	TargetUuid string  `json:"target_uuid"`
	Timestamp  int64   `json:"timestamp"`
	Latency    float64 `json:"latency"`
}

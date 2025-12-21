package model

type Loss struct {
	TargetUuid string `json:"target_uuid"`
	Timestamp  int64  `json:"timestamp"`
}

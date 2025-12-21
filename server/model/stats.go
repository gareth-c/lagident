package model

import "database/sql"

type Stats struct {
	TargetUuid string          `json:"target_uuid"`
	State      string          `json:"state"`
	Sent       uint64          `json:"sent"`
	Recv       uint64          `json:"recv"`
	Last       float64         `json:"last"`
	Loss       float64         `json:"loss"`
	Sum        float64         `json:"sum"`
	Max        float64         `json:"max"`
	Min        sql.NullFloat64 `json:"min"`
	Avg15m     float64         `json:"avg15m"`
	Avg6h      float64         `json:"avg6h"`
	Avg24h     float64         `json:"avg24h"`
	Timestamp  int64           `json:"timestamp"`
}

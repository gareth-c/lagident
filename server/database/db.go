package database

import (
	"database/sql"
	"lagident/model"
	"time"
)

type DB interface {
	GetTechnologies() ([]*model.Technology, error)
	GetTargets() ([]*model.Target, error)
	AddTarget(target model.Target) error
	GetTargetByUuid(uuid string) (*model.Target, error)
	DeleteTarget(uuid string) error
	GetStatsByUuid(uuid string) (*model.Stats, error)
	GetStats() ([]*model.Stats, error)
	SaveStats(stats model.Stats) error
	DeleteStats(uuid string) error
	SaveLoss(loss *model.Loss) error
	DeleteOldLosses(before time.Time) error
	GetLossByUuid(uuid string) ([]model.Loss, error)
	SaveLatency(latency *model.Latency) error
	DeleteOldLatencies(before time.Time) error
	GetLatencyByUuid(uuid string) ([]model.Latency, error)
	SaveMeasurement(m *model.HistogramMeasurement) error
	DeleteOldHistograms(before time.Time) error
	GetHistogramByUuid(uuid string) ([]*model.HistogramMeasurement, error)
}

func NewDB(db *sql.DB, dbType string) DB {
	switch dbType {
	case "mysql":
		return MySQLDB{db: db}
	case "sqlite":
		return SQLiteDB{db: db}
	default:
		panic("Unsupported DB_TYPE. Please set DB_TYPE to 'mysql' or 'sqlite'.")
	}
}

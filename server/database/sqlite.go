package database

import (
	"database/sql"
	"lagident/model"
	"log"
	"time"
)

type SQLiteDB struct {
	db *sql.DB
}

func (d SQLiteDB) GetTechnologies() ([]*model.Technology, error) {
	rows, err := d.db.Query("select name, details from technologies")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tech []*model.Technology
	for rows.Next() {
		t := new(model.Technology)
		err = rows.Scan(&t.Name, &t.Details)
		if err != nil {
			return nil, err
		}
		tech = append(tech, t)
	}
	return tech, nil
}

func (d SQLiteDB) GetTargets() ([]*model.Target, error) {
	rows, err := d.db.Query("SELECT uuid, name, address from targets")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var targets []*model.Target
	for rows.Next() {
		t := new(model.Target)
		err = rows.Scan(&t.Uuid, &t.Name, &t.Address)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (d SQLiteDB) AddTarget(target model.Target) error {
	stmt, err := d.db.Prepare("INSERT INTO targets (uuid, name, address) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(target.Uuid, target.Name, target.Address)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) GetTargetByUuid(uuid string) (*model.Target, error) {
	var target model.Target
	err := d.db.QueryRow("SELECT uuid, name, address FROM targets WHERE uuid = ?", uuid).Scan(&target.Uuid, &target.Name, &target.Address)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No result found
		}
		return nil, err
	}
	return &target, nil
}

func (d SQLiteDB) DeleteTarget(uuid string) error {
	stmt, err := d.db.Prepare("DELETE FROM targets WHERE uuid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) GetStatsByUuid(uuid string) (*model.Stats, error) {
	var stats model.Stats
	err := d.db.QueryRow("SELECT target_uuid, state, sent, recv, last, loss, sum, max, min, avg15m, avg6h, avg24h, timestamp FROM statistics WHERE target_uuid = ?", uuid).Scan(
		&stats.TargetUuid, &stats.State, &stats.Sent, &stats.Recv, &stats.Last, &stats.Loss, &stats.Sum, &stats.Max, &stats.Min, &stats.Avg15m, &stats.Avg6h, &stats.Avg24h, &stats.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No result found
		}
		return nil, err
	}
	return &stats, nil
}

func (d SQLiteDB) GetStats() ([]*model.Stats, error) {
	rows, err := d.db.Query("SELECT target_uuid, state, sent, recv, last, loss, sum, max, min, avg15m, avg6h, avg24h, timestamp FROM statistics")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats []*model.Stats
	for rows.Next() {
		s := new(model.Stats)
		err = rows.Scan(&s.TargetUuid, &s.State, &s.Sent, &s.Recv, &s.Last, &s.Loss, &s.Sum, &s.Max, &s.Min, &s.Avg15m, &s.Avg6h, &s.Avg24h, &s.Timestamp)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (d SQLiteDB) SaveStats(stats model.Stats) error {
	sql := `
    INSERT INTO statistics (target_uuid, state, sent, recv, last, loss, sum, max, min, avg15m, avg6h, avg24h, timestamp)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?)
    ON CONFLICT(target_uuid) DO UPDATE SET
        state = excluded.state,
        sent = excluded.sent,
        recv = excluded.recv,
        last = excluded.last,
        loss = excluded.loss,
        sum = excluded.sum,
        max = excluded.max,
        min = excluded.min,
        avg15m = excluded.avg15m,
        avg6h = excluded.avg6h,
        avg24h = excluded.avg24h,
        timestamp = excluded.timestamp
`
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		stats.TargetUuid, stats.State, stats.Sent, stats.Recv, stats.Last, stats.Loss, stats.Sum, stats.Max, stats.Min, stats.Avg15m, stats.Avg6h, stats.Avg24h, stats.Timestamp,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) DeleteStats(uuid string) error {
	stmt, err := d.db.Prepare("DELETE FROM statistics WHERE target_uuid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(uuid)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) SaveLoss(loss *model.Loss) error {
	sql := "INSERT INTO losses (target_uuid, timestamp) VALUES (?,?)"
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		loss.TargetUuid, loss.Timestamp,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) DeleteOldLosses(before time.Time) error {
	sql := `
    DELETE FROM losses
    WHERE timestamp < ?
    `
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(before.Unix())
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) GetLossByUuid(uuid string) ([]model.Loss, error) {
	rows, err := d.db.Query("SELECT target_uuid, timestamp FROM losses WHERE target_uuid = ?  ORDER BY timestamp ASC", uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var measurements []model.Loss
	for rows.Next() {
		l := new(model.Loss)
		err = rows.Scan(&l.TargetUuid, &l.Timestamp)
		if err != nil {
			return nil, err
		}
		measurements = append(measurements, *l)
	}
	return measurements, nil
}

func (d SQLiteDB) SaveLatency(latency *model.Latency) error {
	sql := "INSERT INTO latencies (target_uuid, timestamp, latency) VALUES (?,?,?)"
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		latency.TargetUuid, latency.Timestamp, latency.Latency,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) DeleteOldLatencies(before time.Time) error {
	sql := `
    DELETE FROM latencies
    WHERE timestamp < ?
    `
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(before.Unix())
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) GetLatencyByUuid(uuid string) ([]model.Latency, error) {
	rows, err := d.db.Query("SELECT target_uuid, timestamp, latency FROM latencies WHERE target_uuid = ?  ORDER BY timestamp ASC", uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var measurements []model.Latency
	for rows.Next() {
		l := new(model.Latency)
		err = rows.Scan(&l.TargetUuid, &l.Timestamp, &l.Latency)
		if err != nil {
			return nil, err
		}
		measurements = append(measurements, *l)
	}
	return measurements, nil
}

func (d SQLiteDB) SaveMeasurement(m *model.HistogramMeasurement) error {
	// We do not need count as it is 1 by default
	sql := `
	INSERT INTO histograms (target_uuid, timestamp, bucket) VALUES (?,?,?)
	ON CONFLICT (target_uuid, timestamp, bucket) DO UPDATE
    SET count = count + 1
	`
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		m.TargetUuid, m.Timestamp, m.Bucket,
	)
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) DeleteOldHistograms(before time.Time) error {
	sql := `
    DELETE FROM histograms
    WHERE timestamp < ?
    `
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(before.Unix())
	if err != nil {
		return err
	}

	return nil
}

func (d SQLiteDB) GetHistogramByUuid(uuid string) ([]*model.HistogramMeasurement, error) {
	rows, err := d.db.Query("SELECT target_uuid, timestamp, bucket, `count` FROM histograms WHERE target_uuid = ?  ORDER BY timestamp, bucket ASC", uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var measurements []*model.HistogramMeasurement
	for rows.Next() {
		m := new(model.HistogramMeasurement)
		err = rows.Scan(&m.TargetUuid, &m.Timestamp, &m.Bucket, &m.Count)
		if err != nil {
			return nil, err
		}
		measurements = append(measurements, m)
	}
	return measurements, nil
}

func InitializeSQLiteDB(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS targets (
            uuid CHAR(36) NOT NULL PRIMARY KEY,
            name TEXT NOT NULL,
            address TEXT NOT NULL
        );`,

		`INSERT OR IGNORE INTO targets VALUES (
            '38c84db2-1c79-40c6-86aa-650474f2cc88', 'localhost', '127.0.0.1'
        );`,

		`CREATE TABLE IF NOT EXISTS statistics (
            target_uuid CHAR(36) NOT NULL PRIMARY KEY,
            state TEXT NOT NULL,
            sent INTEGER NOT NULL DEFAULT 0,
            recv INTEGER NOT NULL DEFAULT 0,
            last REAL DEFAULT 0,
            loss REAL DEFAULT 0,
            sum REAL DEFAULT 0,
            max REAL DEFAULT 0,
            min REAL DEFAULT NULL,
            avg15m REAL DEFAULT 0,
            avg6h REAL DEFAULT 0,
            avg24h REAL DEFAULT 0,
            timestamp INTEGER NOT NULL
        );`,

		`CREATE TABLE IF NOT EXISTS losses (
            target_uuid CHAR(36) NOT NULL,
            timestamp INTEGER NOT NULL,
            PRIMARY KEY (target_uuid, timestamp)
        );`,

		`CREATE TABLE IF NOT EXISTS latencies (
            target_uuid CHAR(36) NOT NULL,
            timestamp INTEGER NOT NULL,
            latency REAL NOT NULL,
            PRIMARY KEY (target_uuid, timestamp)
        );`,

		`CREATE TABLE IF NOT EXISTS histograms (
            target_uuid CHAR(36) NOT NULL,
            timestamp INTEGER NOT NULL,
            bucket REAL DEFAULT 0,
            count INTEGER DEFAULT 1,
            PRIMARY KEY (target_uuid, timestamp, bucket)
        );`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Error executing query: %s\n", query)
			return err
		}
	}

	return nil
}

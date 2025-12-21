package database

import (
	"database/sql"
	"lagident/model"
	"time"
)

type MySQLDB struct {
	db *sql.DB
}

func (d MySQLDB) GetTechnologies() ([]*model.Technology, error) {
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

func (d MySQLDB) GetTargets() ([]*model.Target, error) {
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

func (d MySQLDB) AddTarget(target model.Target) error {
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

func (d MySQLDB) GetTargetByUuid(uuid string) (*model.Target, error) {
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

func (d MySQLDB) DeleteTarget(uuid string) error {
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

func (d MySQLDB) GetStatsByUuid(uuid string) (*model.Stats, error) {
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

func (d MySQLDB) GetStats() ([]*model.Stats, error) {
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

func (d MySQLDB) SaveStats(stats model.Stats) error {
	sql := `
	INSERT INTO statistics (target_uuid, state, sent, recv, last, loss, sum, max, min, avg15m, avg6h, avg24h, timestamp) VALUES (?,?,?,?,?,?,?,?,NULLIF(?, ''),?,?,?,?)
	ON DUPLICATE KEY UPDATE state=VALUES(state), sent=VALUES(sent), recv=VALUES(recv), last=VALUES(last), loss=VALUES(loss), sum=VALUES(sum), max=VALUES(max), min=VALUES(min), avg15m=VALUES(avg15m), avg6h=VALUES(avg6h), avg24h=VALUES(avg24h), timestamp=VALUES(timestamp)
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

func (d MySQLDB) DeleteStats(uuid string) error {
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

func (d MySQLDB) SaveLoss(loss *model.Loss) error {
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

func (d MySQLDB) DeleteOldLosses(before time.Time) error {
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

func (d MySQLDB) GetLossByUuid(uuid string) ([]model.Loss, error) {
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

func (d MySQLDB) SaveLatency(latency *model.Latency) error {
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

func (d MySQLDB) DeleteOldLatencies(before time.Time) error {
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

func (d MySQLDB) GetLatencyByUuid(uuid string) ([]model.Latency, error) {
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

func (d MySQLDB) SaveMeasurement(m *model.HistogramMeasurement) error {
	// We do not need count as it is 1 by default
	sql := `
	INSERT INTO histograms (target_uuid, timestamp, bucket) VALUES (?,?,?)
    ON DUPLICATE KEY UPDATE
    count = count + 1
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

func (d MySQLDB) DeleteOldHistograms(before time.Time) error {
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

func (d MySQLDB) GetHistogramByUuid(uuid string) ([]*model.HistogramMeasurement, error) {
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

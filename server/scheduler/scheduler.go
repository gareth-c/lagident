package scheduler

import (
	"context"
	"fmt"
	"lagident/database"
	"lagident/model"
	"math"
	"sync"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

type Factors struct {
	Fac15m float64
	Fac6h  float64
	Fac24h float64
}

type Scheduler struct {
	db       database.DB
	wg       sync.WaitGroup
	reload   chan struct{}
	shutdown chan struct{}
	interval int64
	factors  Factors
}

func NewScheduler(db database.DB) *Scheduler {
	reload := make(chan struct{})
	shutdown := make(chan struct{})

	var interval int64 = 15
	return &Scheduler{
		db:       db,
		reload:   reload,
		shutdown: shutdown,
		interval: interval,
		factors: Factors{
			Fac15m: math.Exp(-float64(interval) / (15 * 60)),
			Fac6h:  math.Exp(-float64(interval) / (6 * 60 * 60)),
			Fac24h: math.Exp(-float64(interval) / (24 * 60 * 60)),
		},
	}
}

func (s *Scheduler) StartScheduler(parent context.Context) {
	timeout := time.Duration(10 * time.Second)
	interval := time.Duration(s.interval) * time.Second

	s.wg.Add(1)
	go func() {

		defer s.wg.Done()

		ctx, cancel := context.WithCancel(parent)
		defer cancel()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run the first ping immediately
		s.runPings(ctx, timeout)

		for {
			select {
			case <-ctx.Done():
				return

			case _, ok := <-s.shutdown:
				if !ok {
					fmt.Println("Scheduler shutdown")
					return
				}

			case <-ticker.C:
				s.runPings(ctx, timeout)
			}
		}

	}()
}

func (s *Scheduler) StopScheduler() {
	close(s.shutdown)
	close(s.reload)

	s.wg.Wait()
}

func (s *Scheduler) runPings(ctx context.Context, timeout time.Duration) error {
	targets, err := s.db.GetTargets()
	if err != nil {
		fmt.Println("Error getting targets", err)
		return err
	}

	if len(targets) == 0 {
		fmt.Println("No targets found")
		return nil
	}

	wg := sync.WaitGroup{}
	for _, target := range targets {
		wg.Add(1)
		go func(target *model.Target) {
			defer wg.Done()

			pinger, err := probing.NewPinger(target.Address)
			if err != nil {
				// Most of the time this happens if we can't resolve the hostname
				fmt.Printf("Error creating pinger for %s: %v\n", target.Address, err)

				dbStats, err := s.db.GetStatsByUuid(target.Uuid)
				if err != nil {
					fmt.Printf("Error getting stats for %s: %v\n", target.Address, err)
					return
				}

				if dbStats == nil {
					// We do not have any stats for this target yet
					dbStats = &model.Stats{
						TargetUuid: target.Uuid,
						Max:        0,
					}
				}

				// Target is down so we do not modify min, max or the buckets
				dbStats.Sent++
				dbStats.Loss++
				dbStats.State = "down"
				dbStats.Timestamp = time.Now().Unix()

				s.db.SaveLoss(&model.Loss{
					TargetUuid: target.Uuid,
					Timestamp:  time.Now().Unix(),
				})

				s.db.SaveStats(*dbStats)
				return
			}

			pinger.Timeout = timeout
			pinger.Count = 1

			pinger.OnFinish = func(stats *probing.Statistics) {
				dbStats, err := s.db.GetStatsByUuid(target.Uuid)
				if err != nil {
					fmt.Printf("Error getting stats for %s: %v\n", target.Address, err)
					return
				}

				//currentLatency := float64(stats.MaxRtt.Milliseconds())

				// Convert MaxRtt to milliseconds with floating point precision
				currentLatency := float64(stats.MaxRtt) / float64(time.Millisecond)

				if dbStats == nil {
					// We do not have any stats for this target yet
					dbStats = &model.Stats{
						TargetUuid: target.Uuid,
						Max:        currentLatency,
					}
				}

				dbStats.Sent++
				dbStats.State = "up"
				if stats.PacketLoss > 0 {
					// Target is down
					dbStats.Loss++
					dbStats.State = "down"

					err = s.db.SaveLoss(&model.Loss{
						TargetUuid: target.Uuid,
						Timestamp:  time.Now().Unix(),
					})
					if err != nil {
						fmt.Printf("Error saving loss for %s: %v\n", target.Address, err)
					}

				} else {
					dbStats.Recv++
				}

				min := currentLatency
				if dbStats.Min.Valid && currentLatency > 0 {
					min = math.Min(dbStats.Min.Float64, currentLatency)
				} else if dbStats.Min.Valid && dbStats.Min.Float64 > 0 && currentLatency == 0 {
					min = dbStats.Min.Float64
				}

				// Basically this is a Go version of of the original meshping code
				// by Michael Ziegler (Svedrin)
				// https://github.com/Svedrin/meshping/blob/8f6334ab3c362531be6c43fdad67ec321daa2d18/src/meshping.py#L199-L213
				// He is my brother, so I guess it's ok to steal it
				// (👉ﾟヮﾟ)👉
				dbStats.Last = currentLatency
				dbStats.Sum += dbStats.Last
				dbStats.Max = math.Max(dbStats.Max, currentLatency)
				dbStats.Min.Scan(min)
				dbStats.Avg15m = s.expAvg(dbStats.Avg15m, currentLatency, s.factors.Fac15m)
				dbStats.Avg6h = s.expAvg(dbStats.Avg6h, currentLatency, s.factors.Fac6h)
				dbStats.Avg24h = s.expAvg(dbStats.Avg24h, currentLatency, s.factors.Fac24h)
				dbStats.Timestamp = time.Now().Unix()

				err = s.db.SaveStats(*dbStats)
				if err != nil {
					fmt.Printf("Error saving stats for %s: %v\n", target.Address, err)
				}

				if dbStats.State == "up" {
					s.db.SaveLatency(&model.Latency{
						TargetUuid: target.Uuid,
						Timestamp:  time.Now().Unix(),
						Latency:    currentLatency,
					})

					// The plan is to use eCharts to display the histogram
					// intead of the original meshping implementation I simplified this
					// Original would be: int64(math.Log2(currentLatency) * 10)
					//
					// I on the other hand just use the last two digits of the latency to create the bucket

					s.db.SaveMeasurement(&model.HistogramMeasurement{
						TargetUuid: target.Uuid,
						Timestamp:  int64(time.Now().Unix()/3600) * 3600,
						Bucket:     roundFloat(currentLatency, 2.),
					})
				}

			}

			err = pinger.RunWithContext(ctx)
			if err != nil {
				fmt.Printf("Error running pinger for %s: %v\n", target.Address, err)
				return
			}

		}(target)
	}

	return nil
}

func (s *Scheduler) expAvg(current_avg, new_value, factor float64) float64 {
	return (current_avg * factor) + (new_value * (1 - factor))
}

func roundFloat(value float64, precision float64) float64 {
	ratio := math.Pow(10, precision)
	return math.Round(value*ratio) / ratio
}

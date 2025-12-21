package database

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Housekeeping struct {
	db       DB
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewHousekeeping(db DB) *Housekeeping {
	shutdown := make(chan struct{})

	h := &Housekeeping{
		wg:       sync.WaitGroup{},
		db:       db,
		shutdown: shutdown,
	}

	return h
}

func (h *Housekeeping) Start(parent context.Context) {
	h.wg.Add(1)
	go func() {

		defer h.wg.Done()

		ctx, cancel := context.WithCancel(parent)
		defer cancel()

		ticker := time.NewTicker(15 * time.Minute)
		defer ticker.Stop()

		// Run the first ping immediately

		for {
			select {
			case <-ctx.Done():
				return

			case _, ok := <-h.shutdown:
				if !ok {
					fmt.Println("Housekeeping shutdown")
					return
				}

			case <-ticker.C:
				fmt.Println("Do database housekeeping")

				now := time.Now()
				// Subtract 3 days from now
				before := now.AddDate(0, 0, -3)

				h.db.DeleteOldLatencies(before)
				h.db.DeleteOldLosses(before)
				h.db.DeleteOldHistograms(before)
			}
		}

	}()
}

func (h *Housekeeping) StopHousekeeping() {
	close(h.shutdown)

	h.wg.Wait()
}

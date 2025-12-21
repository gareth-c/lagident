package web

import (
	"context"
	"fmt"
	"lagident/database"
	"net/http"
	"os"
	"sync"
	"time"

	"lagident/model"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

type Webserver struct {
	db     database.DB
	wg     sync.WaitGroup
	server *http.Server
	router *gin.Engine
}

type StatisticResponse struct {
	Target     model.Target
	Statistics model.Stats
}

type TimeseriesResponse struct {
	Target    model.Target
	Latencies []model.Latency
	Losses    []model.Loss
}

func NewWebserver(db database.DB, cors bool) *Webserver {
	if os.Getenv("PROFILE") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	webserver := &Webserver{
		wg:     sync.WaitGroup{},
		db:     db,
		server: nil,
		router: gin.Default(),
	}

	webserver.router.Use(disableCors)

	// Serve static files
	webserver.router.Use(static.Serve("/", static.LocalFile("/webapp", true)))

	// API routes
	api := webserver.router.Group("/api")
	if !cors {
		api.Use(disableCors)
	}
	{
		api.GET("/targets", webserver.GetTargets)
		api.GET("/targets/:uuid", webserver.GetTargetByUuid)
		api.POST("/targets/add", webserver.AddTarget)
		api.DELETE("/targets/:uuid", webserver.DeleteTarget)

		api.GET("/statistics", webserver.GetStatistics)

		api.GET("timeseries/:uuid", webserver.GetTimeSeries)
		api.GET("histograms/:uuid", webserver.GetHistogram)
	}

	webserver.server = &http.Server{
		Addr:    ":8080",
		Handler: webserver.router.Handler(),
	}

	return webserver
}

func (w *Webserver) StopWebserver() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	completed := make(chan struct{})

	timeout := time.NewTimer(time.Second * 5)
	defer timeout.Stop()

	go func() {
		_ = w.server.Shutdown(ctx)
		completed <- struct{}{}
	}()

	for {
		select {
		case <-completed:
			return
		case <-timeout.C:
			fmt.Println("web server shutdown reached timeout")
			return
		}
	}

}

func (w *Webserver) StartWebserver(parent context.Context) {
	w.wg.Add(1)

	go func() {
		if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error while starting web server: %v\n", err)
		}
		w.wg.Done()
	}()
}

func disableCors(c *gin.Context) {
	// Source: https://asanchez.dev/blog/cors-golang-options/

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	//c.Header("Content-Type", "application/json")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}

func (w *Webserver) GetTargets(c *gin.Context) {
	targets, err := w.db.GetTargets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, targets)
}

func (w *Webserver) AddTarget(c *gin.Context) {
	var target model.Target
	if err := c.ShouldBindJSON(&target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := w.db.AddTarget(target)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target added successfully"})
}

func (w *Webserver) GetTargetByUuid(c *gin.Context) {
	uuid := c.Param("uuid")
	target, err := w.db.GetTargetByUuid(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if target == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target not found"})
		return
	}
	c.JSON(http.StatusOK, target)
}

func (w *Webserver) DeleteTarget(c *gin.Context) {
	uuid := c.Param("uuid")
	err := w.db.DeleteTarget(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = w.db.DeleteStats(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Target deleted successfully"})
}

func (w *Webserver) GetStatistics(c *gin.Context) {
	targsts, err := w.db.GetTargets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	stats, err := w.db.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	targetMap := make(map[string]model.Target)
	statsMap := make(map[string]model.Stats)

	for _, target := range targsts {
		targetMap[target.Uuid] = *target
	}

	for _, stat := range stats {
		statsMap[stat.TargetUuid] = *stat
	}

	result := make([]StatisticResponse, 0, len(targsts))
	for _, target := range targsts {
		stat, ok := statsMap[target.Uuid]
		if !ok {
			stat = model.Stats{TargetUuid: target.Uuid}
		}
		result = append(result, StatisticResponse{Target: *target, Statistics: stat})
	}

	c.JSON(http.StatusOK, gin.H{"targets": result})
}

func (w *Webserver) GetHistogram(c *gin.Context) {
	uuid := c.Param("uuid")
	histogram, err := w.db.GetHistogramByUuid(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([][]interface{}, 0, len(histogram))
	for _, h := range histogram {
		iso8601Time := time.Unix(h.Timestamp, 0).Format(time.RFC3339)
		response = append(response, []interface{}{iso8601Time, h.Bucket, h.Count})
	}

	c.JSON(http.StatusOK, gin.H{"buckets": response})
}

func (w *Webserver) GetTimeSeries(c *gin.Context) {
	uuid := c.Param("uuid")

	target, err := w.db.GetTargetByUuid(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	latency, err := w.db.GetLatencyByUuid(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	loss, err := w.db.GetLossByUuid(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := TimeseriesResponse{
		Target:    *target,
		Latencies: latency,
		Losses:    loss,
	}

	// Make sure to return an empty array to keep the API consistent
	if response.Latencies == nil {
		response.Latencies = make([]model.Latency, 0)
	}

	if response.Losses == nil {
		response.Losses = make([]model.Loss, 0)
	}

	c.JSON(http.StatusOK, gin.H{"response": response})

}

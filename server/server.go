package main

import (
	"context"
	"database/sql"
	"fmt"
	"lagident/database"
	"lagident/scheduler"
	"lagident/web"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Bind signal handler for linux kernel signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown := make(chan struct{})

	// Select the database type based on an environment variable
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		fmt.Println("DB_TYPE is not set. Defaulting to mysql.")
		dbType = "mysql"
	}

	var d *sql.DB
	var err error

	switch dbType {
	case "mysql":
		d, err = sql.Open("mysql", dataSource())
	case "sqlite":
		d, err = sql.Open("sqlite3", sqlitePath())
		if err == nil {
			err = database.InitializeSQLiteDB(d)
		}
	default:
		log.Fatal("Unsupported DB_TYPE. Please set DB_TYPE to 'mysql' or 'sqlite'.")
	}

	if err != nil {
		log.Fatal(err)
	}

	// CORS is enabled only in prod profile
	cors := os.Getenv("PROFILE") == "prod"

	for {
		fmt.Println("Start Lagident")

		db := database.NewDB(d, dbType)
		go Run(ctx, shutdown, db, cors)

		select {
		case <-ctx.Done():
			return

		case sig := <-sigs:
			if sig.String() == "hangup" {
				log.Println("Start reload of Lagident")

				// Stop the Run() function but do not exit the program.
				// This will trigger a reload because the outer for loop will call the Run() function again.
				shutdown <- struct{}{}
			} else {
				// Stop Process by returning from this function (MainThreadLoop)
				log.Printf("Catch signal: %v - %v", sig, sig.String())
				return
			}
		}
	}
}

func dataSource() string {
	host := os.Getenv("DB_HOST")
	if host == "" {
		if os.Getenv("PROFILE") == "prod" {
			host = "db"
		} else {
			host = "localhost"
		}
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "lagident"
	}

	pass := os.Getenv("DB_PASS")
	if pass == "" {
		pass = "pass"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "lagident"
	}

	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + dbName
}

func sqlitePath() string {
	if os.Getenv("PROFILE") == "prod" {
		return "/data/lagident.db"
	}

	return "lagident.db"
}

func Run(parent context.Context, shutdown chan struct{}, db database.DB, cors bool) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	webserver := web.NewWebserver(db, cors)
	webserver.StartWebserver(ctx)

	scheduler := scheduler.NewScheduler(db)
	scheduler.StartScheduler(ctx)

	housekeeping := database.NewHousekeeping(db)
	housekeeping.Start(ctx)

	select {
	case <-ctx.Done():
	case <-shutdown:
		webserver.StopWebserver()

		scheduler.StopScheduler()

		housekeeping.StopHousekeeping()

		// Give the linux kernel a chance to clean up the socket
		time.Sleep(1 * time.Second)
		return
	}
}

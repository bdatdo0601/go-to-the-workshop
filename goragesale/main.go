package main

import (
	"context"
	"encoding/json" // json package
	"flag"
	"log"      // logging package
	"net/http" // go http server package
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/bdatdo0601/goragesale/schema"
	// "io"       // input/output stream package
)

// ServerErrorString error string
const ServerErrorString = "me boy running to some error"

// Product is an item we sell.
type Product struct {
	ID       string `db:"product_id"`
	Name     string `db:"name"`
	Cost     int    `db:"cost"`
	Quantity int    `db:"quantity"`
}

// Service holds business logic related to Products.
type Service struct {
	db *sqlx.DB
}

func main() {
	// parsing flag from cmd line
	flag.Parse()

	// initialize server
	var db *sqlx.DB
	{
		// set metadata
		q := url.Values{}
		q.Set("sslmode", "disable")
		q.Set("timezone", "utc")

		// set db url
		u := url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword("postgres", "postgres"),
			Host:     "localhost",
			Path:     "postgres",
			RawQuery: q.Encode(),
		}

		// open and create connection to  db
		var err error
		db, err = sqlx.Open("postgres", u.String())
		if err != nil {
			log.Fatalf("error: connecting to db: %s", err)
		}

		defer db.Close()
	}

	// seed and migrate data
	switch flag.Arg(0) {
	case "migrate":
		if err := schema.Migrate(db.DB); err != nil {
			log.Println("error applying migrations", err)
			os.Exit(1)
		}
		log.Println("Migrations complete")
		return

	case "seed":
		if err := schema.Seed(db.DB); err != nil {
			log.Println("error seeding database", err)
			os.Exit(1)
		}
		log.Println("Seed data complete")
		return
	}

	// define appContext
	service := Service{db: db}

	// define port
	const PORT = ":8000"

	// initialize server type
	server := http.Server{
		Addr:         PORT,
		Handler:      http.HandlerFunc(service.ListProducts),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// error storing variable
	// chan is channel
	serverErrors := make(chan error, 1)
	// go routine anonymous function to start server
	// go keyword is used for concurrency
	go func() {
		log.Printf("me boy is starting at %s", server.Addr)
		// channel chaining to error variable
		serverErrors <- server.ListenAndServe()
	}()

	// os signal storing variable with os.Signal channel
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	select {
	// if receive an error, log it
	case err := <-serverErrors:
		log.Fatalf(ServerErrorString+" %s", err)

	case <-osSignals:
		log.Printf("me boy see shutdown signal")

		// give request some timeout
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("me boy cannot gracefully shutting down")
			if err := server.Close(); err != nil {
				log.Printf("error: me boy cannot close: %s", err)
			}
		}

		log.Fatalf("me boy has gracefully shut down")
	}
}

// ListProducts is an HTTP handler for list product endpoints
// Extend from type service
func (appContext *Service) ListProducts(resWriter http.ResponseWriter, req *http.Request) {
	var products []Product

	// retrieve data from db
	if err := appContext.db.Select(&products, "SELECT * FROM products"); err != nil {
		log.Printf("error: selecting products: %s", err)
		resWriter.WriteHeader(http.StatusInternalServerError)
		return
	}
	// setting header for response
	resWriter.Header().Set("Content-Type", "application/json; charset=utf8")

	// encode data and stream it to response writer
	// throw error if something go wrong
	if err := json.NewEncoder(resWriter).Encode(products); err != nil {
		log.Fatalf(ServerErrorString)
	}
}

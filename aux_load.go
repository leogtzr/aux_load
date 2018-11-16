package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// InputFileInfo ...
type InputFileInfo struct {
	ReadyFile string
	ZipFile   string
	LoadFile  string
}

func processFiles(server *http.Server) {
	for i := 0; i < 5; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}

	server.Shutdown(context.Background())
}

func main() {

	// TODO: Get current working dir.
	// TODO: Set up log file:

	m := http.NewServeMux()
	server := http.Server{Addr: ":8000", Handler: m}

	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bye!")
		server.Shutdown(context.Background())
	})

	m.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// stats
	})

	m.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		// stop
	})

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, ":)")
	})

	// main process ...
	go processFiles(&server)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Printf("Finished")
}

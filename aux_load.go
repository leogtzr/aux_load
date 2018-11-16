package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

func processFiles() {
	for i := 0; i < 30; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	m := http.NewServeMux()
	s := http.Server{Addr: ":8000", Handler: m}

	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bye!")
		s.Shutdown(context.Background())
	})

	m.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		// stats
	})

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, ":)")
	})

	// main process ...
	go processFiles()

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Printf("Finished")
}

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kardianos/osext"
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

// func settingUpServer()

func main() {

	// Getting working directory:
	workingDir, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}

	// Setting up logging:
	f, err := os.OpenFile(workingDir+"/aux_load.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	defer f.Close()

	// TODO: Implement CLI options to get the port.
	port := flag.String("host", ":8000", "the port of the application")
	flag.Parse()

	m := http.NewServeMux()
	server := http.Server{Addr: *port, Handler: m}

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
		fmt.Fprintf(w, "OK")
	})

	// main process ...
	go processFiles(&server)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Printf("Finished")
}

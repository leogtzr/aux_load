package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kardianos/osext"
)

const (
	// ConfFileName ...
	ConfFileName            = "env.conf"
	controlFileNotSpecified = "_._notspecified_._"
)

// InputFileInfo ...
type InputFileInfo struct {
	ReadyFile string
	ZipFile   string
	LoadFile  string
}

// Config ...
type Config struct {
	StopFileName string `json:"stopFileName"`
	OnFailEmail  string `json:"onFailEmail"`
	CutOffTime   int    `json:"cutofftime"`
}

func (config *Config) String() string {
	return fmt.Sprintf("StopFileName=%s, OnFailEmail=%q",
		config.StopFileName, config.OnFailEmail)
}

// exists returns whether the given file or directory exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func processFiles(server *http.Server, config *Config, workingDir string) {

	// check to see if the program was forced to stop with the 'stop.txt'
	if exists, _ := exists(config.StopFileName); exists {
		log.Printf("%s file found, stopping	process.", workingDir+"/"+config.StopFileName)
		// Send email
		server.Shutdown(context.Background())
		return
	}

	for i := 0; i < 2; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}

	server.Shutdown(context.Background())
}

func statsHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// stats
	})
}

func stopHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// stop
	})
}

func okHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})
}

func settingUpServer(addr string) *http.Server {
	m := http.NewServeMux()
	server := &http.Server{Addr: addr, Handler: m}

	m.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Bye!")
		server.Shutdown(context.Background())
	})

	m.HandleFunc("/stats", statsHandler())
	m.HandleFunc("/stop", stopHandler())
	m.HandleFunc("/", okHandler())

	return server
}

func readConfigurationFile(path string) (*Config, error) {
	exists, err := exists(path)
	if !exists {
		return nil, fmt.Errorf("'%s' does not exists", path)
	}

	configFile, err := os.Open(path)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}

	var config Config

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

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

	// read configuration file:
	config, err := readConfigurationFile(workingDir + "/" + ConfFileName)
	if err != nil {
		log.Fatal(err)
	}

	port := flag.String("host", ":8000", "the port of the application")
	controlFile := flag.String("ctl", controlFileNotSpecified, "control file to load")
	flag.Parse()

	if *controlFile == controlFileNotSpecified {
		log.Fatal("control file with -ctl option not specified")
	}

	log.Printf("Listening at: %q", *port)

	// Getting the server:
	server := settingUpServer(*port)

	// main process ...
	go processFiles(server, config, workingDir)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	log.Printf("Finished")
}

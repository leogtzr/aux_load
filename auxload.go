package auxload

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
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

// Config represents the basic structure of the configuration used to run this job.
type Config struct {
	StopFileName string `json:"stopFileName"`
	OnFailEmail  string `json:"onFailEmail"`
	CutOffTime   int    `json:"cutOffTime"`
	ControlFile  string
}

// Stats represents the basic of the structure to save information about the files that have
// been already processed.
type Stats struct{}

func (config *Config) String() string {
	return fmt.Sprintf("StopFileName=%s, OnFailEmail=%q, cutOffTime=%q",
		config.StopFileName, config.OnFailEmail, config.CutOffTime)
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

	// main process:
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

func stopHandler(server *http.Server) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// print useful information here.
		server.Shutdown(context.Background())
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
	m.HandleFunc("/stop", stopHandler(server))
	m.HandleFunc("/", okHandler())

	return server
}

func readConfigurationFile(path, controlFile string) (*Config, error) {
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

	config.ControlFile = controlFile
	return &config, nil
}

// Start is the main entry point.
func Start(workingDir, controlFile, addr string) error {
	// Setting up logging:
	f, err := os.OpenFile(workingDir+"/auxload.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	defer f.Close()

	// read configuration file:
	config, err := readConfigurationFile(workingDir+"/"+ConfFileName, controlFile)
	if err != nil {
		return err
	}

	log.Printf("Listening at: %q", addr)

	// Getting the server:
	server := settingUpServer(addr)

	// main process ...
	go processFiles(server, config, workingDir)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	log.Printf("Finished")

	return nil
}

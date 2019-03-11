package auxload

// readonly CHECK_AUX_SCRIPT='jobs/check_aux_db/check_aux_db.sh'

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
	"path/filepath"
)

const (
	// ConfFileName ...
	ConfFileName            = "env.conf"
	controlFileNotSpecified = "_._notspecified_._"
	getCurrentSchemaProgram = "get_current_schema.sh"
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
	workingDir   string
}

// Stats represents the basic of the structure to save information about the files that have
// been already processed.
type Stats struct{}

func (config *Config) String() string {
	return fmt.Sprintf("StopFileName=%s, OnFailEmail=%q, CutOffTime=%q, ControlFile=%q, workingDir=%q",
		config.StopFileName, config.OnFailEmail, config.CutOffTime, config.ControlFile, config.workingDir)
}

func exists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
	}
    return true
}

// code to invoke an external program to get the current datasource.
func currentDataSource(workingDir string) (string, error) {
	var cmdOut []byte

	cmdOut, err := exec.Command(workingDir+"/"+getCurrentSchemaProgram, []string{}...).Output()
	if err != nil {
		return "", fmt.Errorf("there was an error running the command: '%q'", err)
	}

	return string(cmdOut), nil
}

func getAuxSchemaLetter(schema string) string {
	return schema[len(schema)-1:]
}

func getOppositeAuxSchema(schema string) string {
	if schema := strings.ToUpper(schema); schema == "A" {
		return "B"
	}
	return "A"
}

func processFiles(server *http.Server, config *Config, workingDir string) {

	fmt.Println(config)

	// check to see if the program was forced to stop with the 'stop.txt'
	if exists := exists(filepath.Join(config.workingDir, config.StopFileName)); exists {
		log.Printf("%s file found, stopping	process.", filepath.Join(config.workingDir, config.StopFileName))
		// TODO: Send email
		server.Shutdown(context.Background())
		return
	}

	if isFound := exists(filepath.Join(config.workingDir, config.ControlFile + ".running")); isFound {
		log.Printf("%q/%q.running file found, stopping process. Aux database load was already running when it tried to start. Need manual intervention.",
			workingDir, config.ControlFile)
		// TODO: Send email
		server.Shutdown(context.Background())
		fmt.Println(":(")
		return
	}

	dataSorce, err := currentDataSource(workingDir)
	if err != nil {
		log.Println("Error trying to get current schema to load to the offline schema.")
		server.Shutdown(context.Background())
		return
	}

	dataSourceSchemaLetter := getAuxSchemaLetter(dataSorce)
	log.Printf("%q is the current schema, %q", dataSorce, dataSourceSchemaLetter)
	schemaToLoad := getOppositeAuxSchema(dataSourceSchemaLetter)
	log.Println("We will load to: " + schemaToLoad)

	// main process:
	for i := 0; i < 10; i++ {
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

func readConfigurationFile(workingDir, confFileName string) (*Config, error) {
	path := workingDir + "/" + confFileName
	exists := exists(path)
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

	config.ControlFile = confFileName
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
	config, err := readConfigurationFile(workingDir, ConfFileName)
	if err != nil {
		return err
	}
	config.ControlFile = controlFile

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

package vlog

import (
	"io"
	"log"
	"os"
)

type Configuration struct {
	Priority    string `json:"priority"`
	Destination string `json:"destination"`
}

// InitConfiguration copy configuration to local config variable and init the system
func InitConfiguration(c Configuration) {
	config = c
	if len(config.Destination) < 2 || config.Destination == "std" {
		// vlog.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
		Init(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		// try to open the given file
		file, err := os.OpenFile(config.Destination, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file ", config.Destination, ": ", err)
		}
		Init(file, file, file, file)
	}

	log.Printf("log: %v\n", config)
}

var (
	config  Configuration
	Trace   *log.Logger = log.New(os.Stdout, "TRACE-: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info    *log.Logger = log.New(os.Stdout, "INFO-: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning *log.Logger = log.New(os.Stderr, "WARN-: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error   *log.Logger = log.New(os.Stderr, "ERR-: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Init(traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

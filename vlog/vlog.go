package vlog

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type section struct {
	Section *configuration `json:"log"`
}
type configuration struct {
	Priority    string `json:"priority"`
	Destination string `json:"destination"`
}

// parse config section and apply the configuration
func InitConfiguration(jb []byte) {
	var s section
	s.Section = &config
	err := json.Unmarshal(jb, &s)
	if err != nil {
		log.Fatal("Config Parse Error:", err)
	}
	log.Printf("log: %v\n", config)
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
}

var (
	config  configuration
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
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

package jsonconfig

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

// Parser must implement ParseJSON
type Parser interface {
	ParseJSON([]byte) error
}

// Load the JSON config file
func Load(configFile string, p Parser) {
	var err error
	var input = io.ReadCloser(os.Stdin)
	if input, err = os.Open(configFile); err != nil {
		log.Fatalln(err)
	}

	// Read the config file
	jsonBytes, err := ioutil.ReadAll(input)
	input.Close()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("before filtering out comments: %v\n", jsonBytes)

	var reComments = regexp.MustCompile(`(?m)^\s*//.*$\n?`)
	b := reComments.ReplaceAll(jsonBytes, []byte{})
	log.Printf("no comments %v", string(b))

	// Parse the config
	if err := p.ParseJSON(b); err != nil {
		log.Fatalf("Could not parse %q: %v", configFile, err)
	}
}

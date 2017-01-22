package vfiles

import (
	"encoding/json"
	"log"

	"github.com/valeriugold/vket/vfiles/vlocal"
)

type section struct {
	Section *configuration `json:"vfiles"`
}
type configuration struct {
	StorageType string `json:"storageType"`
}

// parse config section and apply the configuration
func InitConfiguration(jb []byte) {
	var config configuration
	var s section
	s.Section = &config
	err := json.Unmarshal(jb, &s)
	if err != nil {
		log.Fatal("Config Parse Error:", err)
	}
	log.Printf("vfiles: %v\n", config)
	if config.StorageType == "local" {
		vbox = vlocal.InitConfiguration(jb)
		// } else if config.StorageType == "s3" {
		// 	vbox = vs3.InitConfiguration(jb)
	}
}

var vbox VFilesBox

func FileSave(nameLocal, nameBox string) error {
	return vbox.FileSave(nameLocal, nameBox)
}
func FileGet(nameLocal, nameBox string) error {
	return vbox.FileGet(nameLocal, nameBox)
}
func FileRemove(nameBox string) error {
	return vbox.FileRemove(nameBox)
}

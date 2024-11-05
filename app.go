package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Configuration struct {
	RptDirectory   string
	AARDirectory   string
	ORBATDirectory string
}

const CONFIG_FILE string = "config.json"
const AAR_DIR_NAME string = "aars"
const AAR_CONFIG_FILENAME string = "aarListConfig.ini"
const ORBAT_FILENAME string = "orbat_%s.json"

var configuration *Configuration

func main() {
	// -- Get exe file location
	// getExecutionLocation()

	// -- Read config
	readConfig(CONFIG_FILE)

}

func getExecutionLocation() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

func readConfig(filename string) {
	file, _ := os.Open(filename)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(configuration.RptDirectory)
	fmt.Println(configuration.AARDirectory)
	fmt.Println(configuration.ORBATDirectory)
}

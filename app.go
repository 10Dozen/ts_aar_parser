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

	ParseRPT(configuration.RptDirectory)

	exportOrbat("Mission1_14-03-2024")

	/*

		Parse(`[tS_ORBAT] [""BLUFOR"",""Razor 1'1"",""RED - FTL"",""CORPORAL"",""Nickname1[kek]""]`)
		Parse(`<AAR-cup_chernarus_A334430><421><unit>[10,0,0,0,1,513]</unit></421></AAR-cup_chernarus_A334430>`)
		Parse(`[tS_ORBAT] [""BLUFOR"",""Razor 1'1"",""Automatic Rifleman"",""CORPORAL"",""Nickname2""]`)
		Parse(`<AAR-cup_chernarus_A334430><421><unit>[11,0,0,0,1,513]</unit></421></AAR-cup_chernarus_A334430>`)
		Parse(`[tS_ORBAT] [""BLUFOR"",""Razor 1'2"",""Razor 1'2 Squad Leader"",""SERGEANT"",""Nic3""]`)
		Parse(`<AAR-cup_chernarus_A334430><421><unit>[12,2300,9577,151,0,-1]</unit></421></AAR-cup_chernarus_A334430>`)
		Parse(`[tS_ORBAT] [""BLUFOR"",""Razor 1'1"",""Rifleman"",""PRIVATE"",""Nick444""]`)

		Parse(`12:33:43.934 [tS_ORBAT] ["BLUFOR","Razor 1'2","FTL1","CORPORAL","Nickname"]`)
		Parse(`12:33:43.934 [tS_ORBAT] ["OPFOR","Razor 1'2","Пулеметчик","PRIVATE","Nickname"]`)

		exportOrbat("Mission1_14-03-2024")
	*/

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
	configuration = &Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
}

func exportOrbat(filenameSuffix string) {
	path := filepath.Join(configuration.ORBATDirectory, fmt.Sprintf("orbat_%s.json", filenameSuffix))
	fmt.Printf("Exporting ORBAT to %s\n", path)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(OrbatAsJSON())
	file.Close()

	path = filepath.Join(configuration.ORBATDirectory, fmt.Sprintf("orbat_leaders_%s.json", filenameSuffix))
	fmt.Printf("Exporting ORBAT Leaders to %s\n", path)
	file2, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file2.Close()
	file2.WriteString(OrbatLeadersAsJSON())
}

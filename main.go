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
	ExecDirectory  string
}

const CONFIG_FILE string = "config.json"
const AAR_DIR_NAME string = "aars"
const AAR_CONFIG_FILENAME string = "aarListConfig.ini"
const ORBAT_FILENAME string = "orbat_%s.json"

var configuration *Configuration = new(Configuration)

func main() {
	// -- Get exe file location
	getExecutionLocation()

	// -- Read config
	// readConfig(CONFIG_FILE)

	// -- Parse RPT file and gather ORBAT data and AAR metadata for futher selection
	//    Will also create tmp intemediate files for each AAR that will be used to fully parse AAR if selected.
	//    These files will be deleted afterward
	// ParseRPT(configuration.RptDirectory)

	// -- Export ORBAT
	//exportOrbat("Mission1_14-03-2024")

	defer closeAllTempReports()

	ParseAARLine(`20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""dingor"", ""Name"": ""CO16 Western"", ""guid"": ""dingor82583"", ""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"`)
	//ParseAARLine(`20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""lingor"", ""name"": ""CO16 Eastern"", ""guid"": ""dingor82586"", ""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"`)
	ParseAARLine(`20:15:53 "<AAR-dingor82583><0><unit>[0,8329,2359,168,1,-1]</unit></0></AAR-dingor82583>"`)
	ParseAARLine(`21:04:09 "<AAR-cup_chernarus_A334430><meta><veh>{ ""vehMeta"": [517,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>""`)
	ParseAARLine(`21:04:09 Some body,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>""`)

	v, _ := json.MarshalIndent(aars, "", "    ")
	fmt.Println(string(v))

	// -- Ask user for excluding some aars if present using AAR metadata
	// handleReportSelection()

	// -- Parse
	// ParseAARs()

	// -- Export AARs???

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

func getExecutionLocation() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	configuration.ExecDirectory = filepath.Dir(ex)
}

func readConfig(filename string) {
	file, _ := os.Open(filename)
	defer file.Close()
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}
}

func handleReportSelection() {
	for {
		fmt.Println("Обнаруженные AAR:")

		for idx, guid := range aarsOrder {
			aar := aars[guid]
			str := fmt.Sprintf(
				"%d) %s",
				idx+1,
				fmt.Sprintf("%s > %s > %s (%s)", aar.timelabel, aar.Name, aar.Terrain, aar.Summary),
			)
			if aar.exclude {
				str = str + " [ ИСКЛЮЧЕН ]"
			}
			fmt.Println(str)
		}

		var excludeId int
		fmt.Print("Укажите ID AAR для исключения: ")
		fmt.Scanf("%d\n", &excludeId)
		if excludeId == 0 {
			break
		}
		// -- Exclude logic here
		if excludeId > len(aars) || excludeId < 1 {
			continue
		}
		aars[aarsOrder[excludeId-1]].exclude = !aars[aarsOrder[excludeId-1]].exclude
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

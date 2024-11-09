package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

/*
TODO:
- Update aarListConfig.ini
+ Excluded AAR is exported with 'null' content
+ Clear tmp files for aars
+ Add js prefix to AAR file
+ Clear ORBAT and excluded AAR data when exported/excluded
- Test against JS AAR converter
- Use goroutines?
*/

type Configuration struct {
	RptDirectory   string
	AARDirectory   string
	ORBATDirectory string
	ExecDirectory  string
}

const (
	CONFIG_FILE         string = "config.json"
	AAR_DIR_NAME               = "aars"
	AAR_CONFIG_FILENAME        = "aarListConfig.ini"
	AAR_FILENAME               = "AAR.%s.%s.%s.json"
	ORBAT_FILENAME             = "ORBAT.%s.json"
	AAR_DATA_PREFIX            = "aarFileData = "
)

var (
	configuration *Configuration = new(Configuration)
	orbatHandler  *ORBATHandler  = NewORBATHandler()
	aarHandler    *AARHandler    = NewAARHandler()
)

func main() {

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGINT)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			aarHandler.Clear()
			os.Exit(0)
		case syscall.SIGINT:
			aarHandler.Clear()
			os.Exit(0)
		}
	}()

	// -- Get exe file location
	getExecutionLocation()

	// -- Read config
	readConfig(CONFIG_FILE)

	// -- Parse RPT file and gather ORBAT data and AAR metadata for futher selection
	//    Will also create tmp intemediate files for each AAR that will be used to fully parse AAR if selected.
	//    These files will be deleted afterward
	rptDate := ParseRPT(configuration.RptDirectory, aarHandler, orbatHandler)

	// -- Export ORBAT
	exportOrbat(rptDate)

	// -- Ask user for excluding some aars if present using AAR metadata
	handleReportSelection()

	// -- Parse
	aarHandler.ParseAARs(rptDate)

	// -- Export AARs
	exportAARs(rptDate)

	// -- Update arrListConfig.ini
	// TBD
}

func getExecutionLocation() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	configuration.ExecDirectory = filepath.Dir(ex)
}

func readConfig(filename string) {
	// filename = filepath.Join(configuration.ExecDirectory, filename)
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
		fmt.Println("\n------------------\nОбнаруженные AAR:\n")

		for idx, aar := range aarHandler.aars {
			excludePrefix := ""
			if aar.exclude {
				excludePrefix = "[ ИСКЛЮЧЕН ] "
			}

			fmt.Println(fmt.Sprintf(
				"%d) %s%s",
				idx+1,
				excludePrefix,
				fmt.Sprintf(
					"%s > %s > %s (%s)",
					aar.timelabel,
					aar.Name,
					aar.Terrain,
					aar.Summary,
				),
			))
		}

		var excludeId int
		fmt.Print("\n------------------\nНажмите Entar для конвертации, либо укажите ID AAR для исключения: ")
		fmt.Scanf("%d\n", &excludeId)
		if excludeId == 0 {
			break
		}
		// -- Exclude logic here
		if excludeId > len(aarHandler.aars) || excludeId < 1 {
			continue
		}

		aarHandler.aars[excludeId-1].exclude = !aarHandler.aars[excludeId-1].exclude
	}
}

func exportOrbat(filenameSuffix string) {
	path := filepath.Join(
		configuration.ORBATDirectory,
		fmt.Sprintf(
			ORBAT_FILENAME,
			filenameSuffix,
		),
	)
	fmt.Printf("Exporting ORBAT to %s\n", path)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(orbatHandler.ToJSON())
	orbatHandler.Omit()
}

func exportAARs(filenameSuffix string) {
	for _, aar := range aarHandler.aars {
		if aar.exclude {
			aarHandler.OmitAAR(aar)
			continue
		}
		path := filepath.Join(
			configuration.AARDirectory,
			AAR_DIR_NAME,
			fmt.Sprintf(
				AAR_FILENAME,
				filenameSuffix,
				aar.Terrain,
				aar.Name,
			),
		)
		fmt.Printf("Exporting AAR to %s\n", path)
		file, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		file.WriteString(AAR_DATA_PREFIX + aarHandler.ToJSON(aar))
		aarHandler.OmitAAR(aar)
	}
}

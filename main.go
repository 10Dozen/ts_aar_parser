package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

type Configuration struct {
	RptDirectory   string
	AARDirectory   string
	ORBATDirectory string
	ExecDirectory  string
}

const (
	CONFIG_FILE           string = "config.json"
	AAR_DIR_NAME                 = "aars"
	AAR_CONFIG_FILENAME          = "aarListConfig.ini"
	AAR_LINK_TEMPLATE            = "%s/%s"
	AAR_FILENAME_TEMPLATE        = "AAR.%s.%s.%s"
	ORBAT_FILENAME               = "ORBAT.%s.json"
	AAR_DATA_PREFIX              = "aarFileData = "
)

var (
	configuration         *Configuration    = new(Configuration)
	windowsFsRestrictedRE *regexp.Regexp    = regexp.MustCompile(`[\s:*?<>|\\/"]`)
	RegexpRepo            *RegexpRepository = NewRegexRepo()
)

func main() {
	/*
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
	*/

	fmt.Println("       ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
	fmt.Println("       ┃   tS AAR/ORBAT Converter (v.1.1.0)   ┃")
	fmt.Println("       ┃           by 10Dozen                 ┃")
	fmt.Println("       ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
	fmt.Println(" Убедитесь, что настроены пути до соответствующих директорий в файле config.json!")
	fmt.Println()

	// -- Get exe file location
	getExecutionLocation()

	// -- Read config
	readConfig(CONFIG_FILE)

	// -- Parse RPT file and gather ORBAT data and AAR metadata for futher selection
	//    Will also create tmp intemediate files for each AAR that will be used to fully parse AAR if selected.
	//    These files will be deleted afterward
	rptContent := ParseLatestRPTs(configuration.RptDirectory)

	// -- Ask user for excluding some aars if present using AAR metadata
	handleReportSelection(rptContent)
	fmt.Println()

	// -- Export ORBAT
	exportOrbat(rptContent.date, rptContent.orbats)

	// -- Parse AARs
	aars := ParseAARs(rptContent.date, rptContent.aars)

	// -- Export AARs
	exportAARs(rptContent.date, aars)
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

func handleReportSelection(rptContent *ReportContent) {
	for {
		fmt.Println("------------------\nОбнаруженные ORBAT:\n")
		for _, orbat := range rptContent.orbats {
			fmt.Printf("  - %s\n", orbat.Mission)
		}

		fmt.Println("------------------\nОбнаруженные AAR:\n")

		for idx, aar := range rptContent.aars {
			excludePrefix := ""
			if aar.exclude {
				excludePrefix = "[ ИСКЛЮЧЕН ] "
			}

			fmt.Println(fmt.Sprintf(
				"%d) %s%s",
				idx+1,
				excludePrefix,
				fmt.Sprintf(
					"%s ▸ %s ▸ %s (%s)",
					aar.timelabel,
					aar.Name,
					aar.Terrain,
					aar.Summary,
				),
			))
		}

		var excludeId int
		fmt.Print("\n------------------\nНажмите Enter для конвертации, либо укажите ID AAR для исключения: ")
		fmt.Scanf("%d\n", &excludeId)
		if excludeId == 0 {
			break
		}
		// -- Exclude logic here
		if excludeId > len(rptContent.aars) || excludeId < 1 {
			continue
		}

		rptContent.aars[excludeId-1].exclude = !rptContent.aars[excludeId-1].exclude
	}
}

func exportOrbat(filenameSuffix string, orbats []*ORBAT) {
	path := filepath.Join(
		configuration.ORBATDirectory,
		fmt.Sprintf(
			ORBAT_FILENAME,
			filenameSuffix,
		),
	)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	content, err := json.MarshalIndent(orbats, "", "    ")
	if err != nil {
		log.Panicf("Failed to convert ORBAT to JSON")
	}

	_, err = file.Write(content)
	if err != nil {
		log.Panicf("Failed to export ORBAT to %s", file.Name())
	}

	fmt.Printf("ORBAT экспортирован в %s\n", file.Name())
}

func exportAARs(reportDate string, aars []*AARConverted) {
	configEntries := make([]*AARConfigEntry, 0)
	for _, aar := range aars {
		normalizedName := fmt.Sprintf(
			AAR_FILENAME_TEMPLATE,
			reportDate,
			aar.Metadata.Terrain,
			windowsFsRestrictedRE.ReplaceAllString(aar.Metadata.Name, `_`),
		)
		archiveName := fmt.Sprintf("%s.%s", normalizedName, "zip")

		// -- Create ZIP archive
		zipfile, err := os.Create(filepath.Join(
			configuration.AARDirectory,
			AAR_DIR_NAME,
			archiveName,
		))
		if err != nil {
			panic(err)
		}
		defer zipfile.Close()

		writer := zip.NewWriter(zipfile)
		archived, err := writer.Create(fmt.Sprintf("%s.%s", normalizedName, "json"))
		if err != nil {
			panic(err)
		}
		defer writer.Close()

		content, err := json.Marshal(aar)
		if err != nil {
			log.Fatalf("Failed to export AAR %s", aar.Metadata.Name)
		}

		data := AAR_DATA_PREFIX + string(content)
		if _, err := archived.Write([]byte(data)); err != nil {
			panic(err)
		}

		// -- Update config
		configEntries = append(configEntries, NewAARConfigEntry(
			reportDate,
			aar.Metadata.Name,
			aar.Metadata.Terrain,
			fmt.Sprintf(AAR_LINK_TEMPLATE, AAR_DIR_NAME, archiveName),
		))

	}

	slices.Reverse(configEntries)
	UpdateAARListConfig(
		filepath.Join(configuration.AARDirectory, AAR_CONFIG_FILENAME),
		configEntries,
	)
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	AAR_TEST_META_PATTERN string = `<meta><core>`
	AAR_TEST_PATTERN      string = `<AAR-.*>`

	// 20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""dingor"", ""Name"": ""CO16 Western"", ""guid"": ""dingor82583"", ""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"`)
	AAR_METADATA_PATTERN     string = `(.*) "<AAR-.*><meta><core>(.*)<\/core>`
	AAR_UNIT_META_PATTERN    string = `<AAR-.*><meta><unit>(.*)<\/unit>`
	AAR_VEHICLE_META_PATTERN string = `<AAR-.*><meta><veh>(.*)<\/veh>`

	AAR_UNIT_FRAME_PATTERN    string = `<AAR-.*><(\d+)><unit>(.*)<\/unit>`
	AAR_VEHICLE_FRAME_PATTERN string = `<AAR-.*><(\d+)><veh>(.*)<\/veh>`
	AAR_ATTACK_FRAME_PATTERN  string = `<AAR-.*><(\d+)><av>(.*)<\/av>`
)

type AARMeta struct {
	Name      string `json:"name"`
	Terrain   string `json:"island"`
	Guid      string `json:"guid"`
	Summary   string `json:"summary"`
	timelabel string
	exclude   bool
	tmp       *os.File
}

/*
type AARUnitMeta struct {
	id       int
	Name     string
	Side     string
	IsPlayer bool
}

type AARVehicleMeta struct {
	id   int
	Name string
}

type ARRUnitFrame struct {
	x       int
	y       int
	dir     int
	IsAlive bool
	// ?
}

type ARRVehicleFrame struct {
	x         int
	y         int
	dir       int
	IsAlive   bool
	Owner     int
	CrewCount int
}
*/

type AARRegexpRepo struct {
	test,
	metadata, unitMeta, vehicleMeta,
	unitFrame, vehicleFrame, attackFrame *regexp.Regexp
}

type AARHandler struct {
	aars   []*AARMeta
	regexp *AARRegexpRepo
}

func (ah *AARHandler) ParseLine(line string) {
	// -- Check for AAR line
	isAAR := ah.regexp.test.MatchString(line)
	if !isAAR {
		return
	}

	// -- Check for meta
	matches := ah.regexp.metadata.FindStringSubmatch(line)
	if matches != nil {
		fmt.Println("AAR Metadata match")
		core := strings.ReplaceAll(strings.Trim(matches[2], " "), `""`, `"`)
		aarMeta := &AARMeta{
			timelabel: matches[1],
		}
		if err := json.Unmarshal([]byte(core), aarMeta); err != nil {
			panic(err)
		}
		//aarMeta.timelabel = matches[1]

		ah.createTempReport(aarMeta)
		ah.aars = append(ah.aars, aarMeta)

		return
	}

	fmt.Println("AAR date line match")
	ah.appendToTempReport(line)
}

func (ah *AARHandler) createTempReport(aar *AARMeta) {
	ah.closeTmpReport()
	tmpFilepath := filepath.Join(
		configuration.ExecDirectory,
		fmt.Sprintf("%s.tmp", aar.Guid),
	)
	fmt.Printf("Creating temporary report %s\n", tmpFilepath)
	file, err := os.Create(tmpFilepath)
	if err != nil {
		panic(err)
	}
	aar.tmp = file
}

func (ah *AARHandler) appendToTempReport(line string) {
	if len(ah.aars) == 0 {
		log.Print("[AARHandler] Found AAR data, but failed to find AAR metada. Skipping...")
		return
	}
	aar := ah.aars[len(ah.aars)-1]
	aar.tmp.WriteString(line)
}

func (ah *AARHandler) closeTmpReport() {
	fmt.Println("Closing temporary report")
	if len(ah.aars) == 0 {
		return
	}
	aar := ah.aars[len(ah.aars)-1]
	aar.tmp.Close()
}

func NewAARHandler() *AARHandler {
	h := &AARHandler{
		aars: make([]*AARMeta, 0),
		regexp: &AARRegexpRepo{
			test:     regexp.MustCompile(AAR_TEST_PATTERN),
			metadata: regexp.MustCompile(AAR_METADATA_PATTERN),
		},
	}

	return h
}

/*
20:15:53 "<AAR-dingor82583><0><unit>[0,8329,2359,168,1,-1]</unit></0></AAR-dingor82583>"
21:04:09 "<AAR-cup_chernarus_A334430><meta><veh>{ ""vehMeta"": [517,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[500,3246,8407,317,1,56,-1]</veh></7></AAR-cup_chernarus_A334430>"

20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""dingor"", ""name"": ""CO16 Western"", ""guid"": ""dingor82583"",
""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [0,""Osamich"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [1,""Ka6aH"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [2,""invaderok"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [3,""Smoker"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [4,""Реневал"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [5,""10Dozen"",""blufor"",1] }</unit></meta></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [6,""chek1"",""blufor"",1] }</unit></meta></AAR-dingor82583>"

20:15:53 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [16,"""",""civ"",0] }</unit></meta></AAR-dingor82583>"

20:15:53 "<AAR-dingor82583><0><unit>[0,8329,2359,168,1,-1]</unit></0></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><0><unit>[1,8326,2360,168,1,-1]</unit></0></AAR-dingor82583>"
20:15:53 "<AAR-dingor82583><0><unit>[2,8324,2359,168,1,-1]</unit></0></AAR-dingor82583>"

20:15:59 "<AAR-dingor82583><meta><unit>{ ""unitMeta"": [17,"""",""opfor"",0] }</unit></meta></AAR-dingor82583>"



21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[500,3246,8407,317,1,56,-1]</veh></7></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[501,2301,9549,313,1,-1,-1]</veh></7></AAR-cup_chernarus_A334430>"
21:04:13 "<AAR-cup_chernarus_A334430><7><veh>[502,3141,8417,271,1,53,-1]</veh></7></AAR-cup_chernarus_A334430>"

21:04:09 "<AAR-cup_chernarus_A334430><meta><veh>{ ""vehMeta"": [517,""HEMTT Ammo""] }</veh></meta></AAR-cup_chernarus_A334430>"
*/

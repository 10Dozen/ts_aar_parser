package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AARCoreMeta struct {
	Name      string `json:"name"`
	Terrain   string `json:"island"`
	Guid      string `json:"guid"`
	Summary   string `json:"summary"`
	timelabel string
	exclude   bool
	tmp       *os.File
}

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

const (
	TEST_META_PATTERN string = `<meta><core>`
	TEST_PATTERN      string = `<AAR-([a-zA-Z0-9]*)>`

	// 20:15:53 "<AAR-dingor82583><meta><core>{ ""island"": ""dingor"", ""Name"": ""CO16 Western"", ""guid"": ""dingor82583"", ""summary"": ""Ковбои освобождают свой городок от бандитов"" }</core></meta></AAR-dingor82583>"`)
	METADATA_PATTERN     string = `(.*) "<AAR-.*><meta><core>(.*)<\/core>`
	UNIT_META_PATTERN    string = `<AAR-.*><meta><unit>(.*)<\/unit>`
	VEHICLE_META_PATTERN string = `<AAR-.*><meta><veh>(.*)<\/veh>`

	UNIT_FRAME_PATTERN    string = `<AAR-.*><(\d+)><unit>(.*)<\/unit>`
	VEHICLE_FRAME_PATTERN string = `<AAR-.*><(\d+)><veh>(.*)<\/veh>`
	ATTACK_FRAME_PATTERN  string = `<AAR-.*><(\d+)><av>(.*)<\/av>`
)

var (
	aars                                                      map[string]*AARCoreMeta = make(map[string]*AARCoreMeta, 0)
	aarsOrder                                                 []string
	testMetaRE, testRE, metedataRE, unitMetaRE, vehicleMetaRE *regexp.Regexp
	unitFrameRE, vehicleFrameRE, attackFrameRe                *regexp.Regexp
)

func ParseAARLine(line string) {
	if testMetaRE == nil {
		testRE = regexp.MustCompile(TEST_PATTERN)
		metedataRE = regexp.MustCompile(METADATA_PATTERN)
		// etc
	}

	// -- Check for meta
	matches := metedataRE.FindStringSubmatch(line)
	if matches != nil {
		fmt.Println("Match")
		core := strings.ReplaceAll(strings.Trim(matches[2], " "), `""`, `"`)
		aarMeta := new(AARCoreMeta)
		if err := json.Unmarshal([]byte(core), aarMeta); err != nil {
			panic(err)
		}
		aarMeta.timelabel = matches[1]
		aars[aarMeta.Guid] = aarMeta
		aarsOrder = append(aarsOrder, aarMeta.Guid)

		createTempReport(aarMeta)

		return
	}

	// -- Check for AAR line
	matches = testRE.FindStringSubmatch(line)
	if matches == nil {
		return
	}

	fmt.Println("AAR line match")
	guid := matches[1]
	appendToTempReport(aars[guid], line)
}

func createTempReport(aar *AARCoreMeta) {
	a := filepath.Join(configuration.ExecDirectory, fmt.Sprintf("%s.tmp", aar.Guid))
	fmt.Println(a)
	file, err := os.Create(filepath.Join(configuration.ExecDirectory, fmt.Sprintf("%s.tmp", aar.Guid)))
	if err != nil {
		panic(err)
	}
	aar.tmp = file
}

func appendToTempReport(aar *AARCoreMeta, line string) {
	aar.tmp.WriteString(line)
}

func closeAllTempReports() {
	for _, v := range aars {
		fmt.Printf("Closing %s\n", v.tmp.Name())
		v.tmp.Close()
		defer os.Remove(v.tmp.Name())
	}
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

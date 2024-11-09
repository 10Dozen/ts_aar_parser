package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	TagUnit    string = "unit"
	TagVehicle        = "veh"
	TagAttack         = "av"
)

type AAR struct {
	Guid    string `json:"guid"`
	Terrain string `json:"island"`
	Name    string `json:"name"`
	Summary string `json:"summary"`

	out *AARConverted

	exclude        bool
	timelabel      string
	date           string
	players        []string
	buff           *bufio.Writer
	expectedLength int
	tmp            *os.File
}

type AARConverted struct {
	Metadata *AARMetadata `json:"metadata"`
	Frames   []*AARFrame  `json:"timeline"`
}

type AARMetadata struct {
	Terrain  string      `json:"island"`
	Name     string      `json:"name"`
	Duration int         `json:"time"`
	Date     string      `json:"date"`
	Summary  string      `json:"desc"`
	Players  []*AARData  `json:"players"`
	Objects  *AARObjects `json:"objects"`
}

type AARObjects struct {
	Units    []*AARData `json:"units"`
	Vehicles []*AARData `json:"vehs"`
}

type AARFrame struct {
	Units    []*AARData
	Vehicles []*AARData
	Attacks  []*AARData
}

func (f *AARFrame) MarshalJSON() ([]byte, error) {
	units, err := json.Marshal(f.Units)
	if err != nil {
		panic(err)
	}

	vehs, err := json.Marshal(f.Vehicles)
	if err != nil {
		panic(err)
	}

	attacks, err := json.Marshal(f.Attacks)
	if err != nil {
		panic(err)
	}

	out := fmt.Sprintf("[%s, %s, %s]", units, vehs, attacks)
	return []byte(out), nil
}

type AARMetadataUnit struct {
	Id       int
	Name     string
	Side     string
	IsPlayer int
}

func (u *AARMetadataUnit) UnmarshalJSON(buf []byte) error {
	tmp := []interface{}{&u.Id, &u.Name, &u.Side, &u.IsPlayer}
	if err := json.Unmarshal(buf, &tmp); err != nil {
		return err
	}
	return nil
}

type AARData struct {
	Data string
}

func (e AARData) MarshalJSON() ([]byte, error) {
	// -- Export as raw data without extra quotes
	return []byte(e.Data), nil
}

// Parses AAR data stored in temporary file `aar.tmp` and composes data to `AARConverted` struct.
// `AARConverted` struct is ready to export as JSON.
func (aar *AAR) Parse() {
	log.Printf("[%s] Parsing\n", aar.Guid)
	if aar.exclude {
		log.Printf("[%s] Skipped\n", aar.Guid)
		return
	}

	log.Printf("[%s] Going to read tmp file %s\n", aar.Guid, aar.tmp.Name())

	aar.out = &AARConverted{
		Metadata: &AARMetadata{
			Terrain:  aar.Terrain,
			Name:     aar.Name,
			Duration: 0,
			Date:     aar.date,
			Summary:  aar.Summary,
			Players:  make([]*AARData, 0),
			Objects: &AARObjects{
				Units:    make([]*AARData, 0),
				Vehicles: make([]*AARData, 0),
			},
		},
		Frames: make([]*AARFrame, 0, aar.expectedLength/2),
	}

	file, err := os.Open(aar.tmp.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	aar.tmp = file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		aar.parseLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// -- Update
	aar.out.Metadata.Duration = len(aar.out.Frames) - 1

	file.Close()
	return
}

// Parses single text line and search for objects metadata or frame data.
// Saves data to `out.Metadata` or `out.Frames`
func (aar *AAR) parseLine(line string) {
	//fmt.Printf("[%s] Parsing line: %s\n", aar.Guid, line)

	// -- Check for frame data
	matches := aarHandler.regexp.frame.FindStringSubmatch(line)
	if matches != nil {
		idx, err := strconv.Atoi(matches[1])
		if err != nil {
			panic(err)
		}
		aar.handleFrameData(idx, matches[2], matches[3])
		return
	}

	// --- Check for metadata
	matches = aarHandler.regexp.objectMetadata.FindStringSubmatch(line)
	if matches == nil {
		return
	}
	aar.handleObjectData(
		matches[1],
		strings.TrimSpace(strings.ReplaceAll(matches[3], `""`, `"`)),
	)
}

// Handles object metadata (unit or vehicle) - adds unit/vehice to a list (`out.Metadata.Objects.Units/Vehicles`), saves playable objects into `out.Metadata.Players`
func (aar *AAR) handleObjectData(metadataType, content string) {
	//fmt.Printf("[AAR.handleObjectData]  Type: %s, Value: %s\n", metadataType, content)
	if metadataType == TagVehicle {
		//fmt.Println("[AAR.handleObjectData] Saving vehicle data")
		aar.out.Metadata.Objects.Vehicles = append(
			aar.out.Metadata.Objects.Vehicles,
			&AARData{Data: content},
		)
		return
	}

	unit := &AARMetadataUnit{}
	if err := json.Unmarshal([]byte(content), unit); err != nil {
		panic(err)
	}
	//fmt.Printf("%#v\n", unit)

	// -- If player and not added already -- add to players meta
	if unit.IsPlayer == 1 && !slices.ContainsFunc(
		aar.players,
		func(e string) bool {
			return e == unit.Name
		},
	) {
		aar.players = append(aar.players, unit.Name)
		aar.out.Metadata.Players = append(
			aar.out.Metadata.Players,
			&AARData{Data: fmt.Sprintf(`["%s", "%s"]`, unit.Name, unit.Side)},
		)
	}

	//fmt.Println("[AAR.handleObjectData] Saving unit data")
	aar.out.Metadata.Objects.Units = append(
		aar.out.Metadata.Objects.Units,
		&AARData{Data: content},
	)
}

// Handles frame data and saves to `out.Frames` under given index
func (aar *AAR) handleFrameData(idx int, frameType, data string) {
	//fmt.Printf("[AAR.handleFrameData] Idx=%d, Type: %s, Data: %s \n", idx, frameType, data)

	// -- Extend Frames, but in case of missing log second - refill with empty frame
	if len(aar.out.Frames)-1 < idx {
		diff := idx - (len(aar.out.Frames) - 1)
		//fmt.Printf("Extending Timeline with %d frames as IDX is overrun\n", diff)
		for i := 0; i < diff; i++ {
			aar.out.Frames = append(aar.out.Frames, &AARFrame{
				Units:    make([]*AARData, 0),
				Vehicles: make([]*AARData, 0),
				Attacks:  make([]*AARData, 0),
			})
		}
	}

	// -- Get frame to update
	frame := aar.out.Frames[idx]
	frameData := &AARData{Data: data}

	//fmt.Printf("Frame %d:\n  %#v  \n", idx, frame)
	/*
		for _, b := range frame.Units {
			fmt.Printf("  Units:  %#v  \n", b)
		}
		for _, b := range frame.Vehicles {
			fmt.Printf("  Vehs:  %#v  \n", b)
		}
		for _, b := range frame.Vehicles {
			fmt.Printf("  Avs:  %#v  \n", b)
		}
	*/
	switch frameType {
	case TagUnit:
		frame.Units = append(frame.Units, frameData)
	case TagVehicle:
		frame.Vehicles = append(frame.Vehicles, frameData)
	case TagAttack:
		frame.Attacks = append(frame.Attacks, frameData)
	}

	//fmt.Printf("Frame %d after:\n  %#v  \n", idx, frame)
	/*
		for _, b := range frame.Units {
			fmt.Printf("Units:  %#v  \n", b)
		}
		for _, b := range frame.Vehicles {
			fmt.Printf("Vehs:   %#v  \n", b)
		}
		for _, b := range frame.Attacks {
			fmt.Printf("Avs:    %#v  \n", b)
		}
	*/
}

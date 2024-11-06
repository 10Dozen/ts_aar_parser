package main

import (
	"encoding/json"
	"regexp"
	"slices"
	"strings"
)

type ORBATElement struct {
	Role  string
	Rank  string
	Name  string
	Side  string
	Group string
}

type ORBATUnit struct {
	Role string
	Rank string
	Name string
}

type ORBATGroup struct {
	Name  string
	Units []*ORBATUnit
}

type ORBATSide struct {
	Name   string
	Groups []*ORBATGroup
}

type ORBAT struct {
	Sides []*ORBATSide
}

type Leader struct {
	Group string
	Role  string
	Name  string
}

type LeaderORBAT struct {
	HQ           []*Leader
	SquadLeaders []*Leader
	TeamLeaders  []*Leader
}

// 12:33:43.934 [tS_ORBAT] ["BLUFOR", "FTL", "CORPORAL", "Nickname"]
const patternStr string = `\[tS_ORBAT\] (\[.*\])`
const (
	Private    string = "PRIVATE"
	Corporal          = "CORPORAL"
	Sergeant          = "SERGEANT"
	Lieutenant        = "LIEUTENANT"
)

var (
	pattern *regexp.Regexp
	orbat   *ORBAT
	leaders *LeaderORBAT
)

func Orbat() *ORBAT {
	return orbat
}

func OrbatAsJSON() string {
	outputData, err := json.MarshalIndent(orbat, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(outputData)
}

func OrbatLeadersAsJSON() string {
	composeLeaders()
	outputData, err := json.MarshalIndent(leaders, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(outputData)
}

func ParseORBATLine(line string) {
	//fmt.Printf("[orbat_reader.Parse] Invoked. Line=%s \n", line)
	succes, subline := checkLine(line)
	if !succes {
		return
	}
	orbatElement := parseElement(subline)

	// -- Create ORBAT struct if not yet exists
	if orbat == nil {
		orbat = &ORBAT{
			Sides: make([]*ORBATSide, 0),
		}
	}

	// -- Get existing/add missing Side
	idx := slices.IndexFunc(orbat.Sides, func(s *ORBATSide) bool {
		if s == nil {
			return false
		}
		return s.Name == orbatElement.Side
	})
	var side *ORBATSide
	if idx > -1 {
		side = orbat.Sides[idx]
		//fmt.Println("Found side")
	} else {
		side = &ORBATSide{
			Name:   orbatElement.Side,
			Groups: make([]*ORBATGroup, 0),
		}
		orbat.Sides = append(orbat.Sides, side)
		//fmt.Println("New side")
	}

	// -- Get existing/add missing Group
	idx = slices.IndexFunc(side.Groups, func(s *ORBATGroup) bool {
		if s == nil {
			return false
		}
		return s.Name == orbatElement.Group
	})
	var group *ORBATGroup
	if idx > -1 {
		group = side.Groups[idx]
		//fmt.Println("Found group")
	} else {
		group = &ORBATGroup{
			Name:  orbatElement.Group,
			Units: make([]*ORBATUnit, 0),
		}
		side.Groups = append(side.Groups, group)
		//fmt.Println("New group")
	}

	group.Units = append(group.Units, &ORBATUnit{
		Role: orbatElement.Role,
		Rank: orbatElement.Rank,
		Name: orbatElement.Name,
	})
	//fmt.Println("[orbat_reader.Parse] ----END----")
}

func checkLine(line string) (bool, string) {
	//fmt.Printf("[orbat_reader.checkLine] Invoked. line=%s\n", line)
	if pattern == nil {
		pattern = regexp.MustCompile(patternStr)
	}

	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		//fmt.Println("No match")
		//fmt.Println("[orbat_reader.checkLine] ----END----")
		return false, ""
	}
	/*
		fmt.Println("Matches:")
		fmt.Println(matches)
		fmt.Println(len(matches))
		for _, m := range matches {
			fmt.Println(m)
		}
	*/
	//fmt.Println("[orbat_reader.checkLine] ----END----")
	return true, strings.TrimSpace(matches[1])
}

func parseElement(line string) ORBATElement {
	//fmt.Printf("[orbat_reader.parseElement] Invoked. line=%s\n", line)
	line = strings.Trim(line, "[], ")
	elements := strings.Split(line, ",")
	orbatElement := ORBATElement{
		Side:  strings.Trim(elements[0], `"`),
		Group: strings.Trim(elements[1], `"`),
		Role:  strings.Trim(elements[2], `"`),
		Rank:  strings.Trim(elements[3], `"`),
		Name:  strings.Trim(elements[4], `"`),
	}
	//fmt.Println("[orbat_reader.parseElement] Element:")
	//fmt.Println(orbatElement)
	//fmt.Println("[orbat_reader.parseElement] ----END----")
	return orbatElement
}

func composeLeaders() {
	if leaders != nil {
		return
	}

	leaders = &LeaderORBAT{
		SquadLeaders: make([]*Leader, 0),
		TeamLeaders:  make([]*Leader, 0),
		HQ:           make([]*Leader, 0),
	}

	for _, side := range orbat.Sides {
		for _, group := range side.Groups {
			for _, unit := range group.Units {
				leader := &Leader{
					Role:  unit.Role,
					Name:  unit.Name,
					Group: group.Name,
				}
				switch unit.Rank {
				case Private:
					continue
				case Corporal:
					leaders.TeamLeaders = append(leaders.TeamLeaders, leader)
				case Sergeant:
					leaders.SquadLeaders = append(leaders.SquadLeaders, leader)
				default:
					leaders.SquadLeaders = append(leaders.SquadLeaders, leader)
				}
			}
		}
	}
}

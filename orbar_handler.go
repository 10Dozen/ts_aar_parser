package main

import (
	"encoding/json"
	"fmt"
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

// 12:33:43.934 [tS_ORBAT] ["BLUFOR", "FTL", "CORPORAL", "Nickname"]
const patternStr string = `\[tS_ORBAT\] (\[.*\])`

var pattern *regexp.Regexp
var orbat *ORBAT

func Orbat() *ORBAT {
	return orbat
}

func OrbatAsJSON() string {
	outputData, err := json.Marshal(orbat)
	if err != nil {
		panic(err)
	}
	return string(outputData)
}

func Parse(line string) {
	fmt.Printf("[orbat_reader.Parse] Invoked. Line=%s \n", line)
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
		fmt.Println("Found side")
	} else {
		side = &ORBATSide{
			Name:   orbatElement.Side,
			Groups: make([]*ORBATGroup, 0),
		}
		orbat.Sides = append(orbat.Sides, side)
		fmt.Println("New side")
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
		fmt.Println("Found group")
	} else {
		group = &ORBATGroup{
			Name:  orbatElement.Group,
			Units: make([]*ORBATUnit, 0),
		}
		side.Groups = append(side.Groups, group)
		fmt.Println("New group")
	}

	group.Units = append(group.Units, &ORBATUnit{
		Role: orbatElement.Role,
		Rank: orbatElement.Rank,
		Name: orbatElement.Name,
	})
	fmt.Println("[orbat_reader.Parse] ----END----")
}

func checkLine(line string) (bool, string) {
	fmt.Printf("[orbat_reader.checkLine] Invoked. line=%s\n", line)
	if pattern == nil {
		pattern = regexp.MustCompile(patternStr)
	}

	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		fmt.Println("No match")
		fmt.Println("[orbat_reader.checkLine] ----END----")
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
	fmt.Println("[orbat_reader.checkLine] ----END----")
	return true, strings.TrimSpace(matches[1])
}

func parseElement(line string) ORBATElement {
	fmt.Printf("[orbat_reader.parseElement] Invoked. line=%s\n", line)
	line = strings.Trim(line, "[], ")
	elements := strings.Split(line, ",")
	orbatElement := ORBATElement{
		Side:  strings.Trim(elements[0], `"`),
		Group: strings.Trim(elements[1], `"`),
		Role:  strings.Trim(elements[2], `"`),
		Rank:  strings.Trim(elements[3], `"`),
		Name:  strings.Trim(elements[4], `"`),
	}
	fmt.Println("[orbat_reader.parseElement] Element:")
	fmt.Println(orbatElement)
	fmt.Println("[orbat_reader.parseElement] ----END----")
	return orbatElement
}

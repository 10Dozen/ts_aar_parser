package main

import (
	"encoding/json"
	"log"
	"strings"
)

const (
	Private    string = "PRIVATE"
	Corporal          = "CORPORAL"
	Sergeant          = "SERGEANT"
	Lieutenant        = "LIEUTENANT"
)

type ORBAT struct {
	Mission string
	Leaders *ORBATLeaders
	Sides   map[string]*ORBATSide
}

func (o *ORBAT) MarshalJSON() ([]byte, error) {
	sides := make([]*ORBATSide, 0, len(o.Sides))
	for _, v := range o.Sides {
		sides = append(sides, v)
	}
	out, err := json.Marshal(struct {
		ORBAT
		Sides []*ORBATSide
	}{ORBAT: *o, Sides: sides})
	if err != nil {
		panic(err)
	}
	return out, nil
}

type ORBATLeaders struct {
	HQ           []*ORBATLeader
	SquadLeaders []*ORBATLeader
	TeamLeaders  []*ORBATLeader
}

type ORBATSide struct {
	Name   string
	Groups map[string]*ORBATGroup
}

func (s *ORBATSide) MarshalJSON() ([]byte, error) {
	groups := make([]*ORBATGroup, 0, len(s.Groups))
	for _, v := range s.Groups {
		groups = append(groups, v)
	}

	out, err := json.Marshal(struct {
		ORBATSide
		Groups []*ORBATGroup
	}{ORBATSide: *s, Groups: groups})
	if err != nil {
		panic(err)
	}
	return out, nil
}

type ORBATGroup struct {
	Name  string
	Units []*ORBATUnit
}

type ORBATLeader struct {
	Group string
	Role  string
	Name  string
}

type ORBATUnit struct {
	Role  string
	Rank  string
	Name  string
	side  string
	group string
}

type ORBATHandler struct {
	orbats []*ORBAT
}

func (oh *ORBATHandler) ParseLine(line string) {
	// -- Check for ORBAT Metadata
	matches := RegexpRepo.ORBAT.metadataRE.FindStringSubmatch(line)
	if matches != nil {
		missionName := matches[1]
		orbat := &ORBAT{
			Mission: missionName,
			Leaders: &ORBATLeaders{
				HQ:           make([]*ORBATLeader, 0),
				SquadLeaders: make([]*ORBATLeader, 0),
				TeamLeaders:  make([]*ORBATLeader, 0),
			},
			Sides: make(map[string]*ORBATSide, 0),
		}

		if oh.orbats == nil {
			oh.orbats = make([]*ORBAT, 1)
			oh.orbats[0] = orbat
			return
		}

		oh.orbats = append(oh.orbats, orbat)
	}

	// -- Check for ORBAT data line
	matches = RegexpRepo.ORBAT.dataRE.FindStringSubmatch(line)
	if matches == nil || len(matches) < 2 {
		return
	}

	if len(oh.orbats) == 0 {
		log.Println("[ORBAT Handler] Found ORBAT data lines, but there were no ORBAT Metadata yet")
		return
	}
	unit := oh.parseUnit(matches[1])
	orbat := oh.orbats[len(oh.orbats)-1]
	oh.addUnit(unit, orbat)
}

func (oh *ORBATHandler) parseUnit(line string) ORBATUnit {
	elements := make([]string, 0, 5)
	if err := json.Unmarshal(
		[]byte(strings.ReplaceAll(line, `""`, `"`)),
		&elements,
	); err != nil {
		panic(err)
	}

	return ORBATUnit{
		side:  elements[0],
		group: elements[1],
		Role:  elements[2],
		Rank:  elements[3],
		Name:  elements[4],
	}
}

func (oh *ORBATHandler) addUnit(unit ORBATUnit, orbat *ORBAT) {
	side, ok := orbat.Sides[unit.side]
	if !ok {
		side = &ORBATSide{
			Name:   unit.side,
			Groups: make(map[string]*ORBATGroup, 0),
		}
		orbat.Sides[unit.side] = side
	}

	group, ok := side.Groups[unit.group]
	if !ok {
		group = &ORBATGroup{
			Name:  unit.group,
			Units: make([]*ORBATUnit, 0),
		}
		side.Groups[unit.group] = group
	}
	group.Units = append(group.Units, &unit)

	// -- Add leaders if rank is above Private
	leader := &ORBATLeader{
		Role:  unit.Role,
		Name:  unit.Name,
		Group: group.Name,
	}
	switch unit.Rank {
	case Private:
		return
	case Corporal:
		orbat.Leaders.TeamLeaders = append(
			orbat.Leaders.TeamLeaders,
			leader,
		)
	case Sergeant:
		orbat.Leaders.SquadLeaders = append(
			orbat.Leaders.SquadLeaders,
			leader,
		)
	default:
		orbat.Leaders.HQ = append(
			orbat.Leaders.HQ,
			leader,
		)
	}
}

func NewORBATHandler() *ORBATHandler {
	h := &ORBATHandler{
		orbats: make([]*ORBAT, 0),
	}

	return h
}

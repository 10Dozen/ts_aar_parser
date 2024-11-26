package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	FLUSH_AFTER int = 10000
)

type AARHandler struct {
	aars []*AAR
}

type AARConfigEntry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Terrain string `json:"terrain"`
	Link    string `json:"link"`
}

func (ah *AARHandler) ParseLine(line string) {
	// -- Check for AAR line
	if !RegexpRepo.AAR.test.MatchString(line) {
		return
	}

	// -- Check for meta
	if RegexpRepo.AAR.testMeta.MatchString(line) {
		matches := RegexpRepo.AAR.metadata.FindStringSubmatch(line)
		core := strings.ReplaceAll(strings.Trim(matches[2], " "), `""`, `"`)
		aar := &AAR{
			timelabel: matches[1],
		}
		if err := json.Unmarshal([]byte(core), aar); err != nil {
			panic(err)
		}

		ah.createTempReport(aar)
		ah.aars = append(ah.aars, aar)
		return
	}

	ah.appendToTempReport(line)
}

func (ah *AARHandler) createTempReport(aar *AAR) {
	ah.closeTmpReport()
	tmpFilepath := filepath.Join(
		configuration.ExecDirectory,
		fmt.Sprintf("%s.tmp", aar.Guid),
	)

	file, err := os.Create(tmpFilepath)
	if err != nil {
		panic(err)
	}
	aar.buff = bufio.NewWriter(file)
	aar.tmp = file
}

func (ah *AARHandler) appendToTempReport(line string) {
	if len(ah.aars) == 0 {
		log.Print("[AARHandler] Found AAR data, but failed to find AAR metada. Skipping...")
		return
	}
	aar := ah.aars[len(ah.aars)-1]

	aar.expectedLength += 1
	aar.buff.WriteString(line + "\n")
	if aar.expectedLength%FLUSH_AFTER == 0 {
		aar.buff.Flush()
	}
}

func (ah *AARHandler) closeTmpReport() {
	if len(ah.aars) == 0 {
		return
	}

	aar := ah.aars[len(ah.aars)-1]
	if aar.buff == nil {
		return
	}

	// Force current buffer to flush
	aar.buff.Flush()
	aar.buff = nil
	aar.tmp.Close()
}

func (ah *AARHandler) ToJSON(aar *AAR) string {
	outputData, err := json.Marshal(aar.out)
	if err != nil {
		panic(err)
	}
	return string(outputData)
}

func (ah *AARHandler) Clear() {
	for _, aar := range ah.aars {
		ah.DiscardAAR(aar)
	}
}

func (ah *AARHandler) DiscardAAR(aar *AAR) {
	aar.tmp.Close()
	os.Remove(aar.tmp.Name())
	aar = nil
}

func NewAARHandler() *AARHandler {
	h := &AARHandler{
		aars: make([]*AAR, 0),
	}

	return h
}

func NewAARConfigEntry(date, title, terrain, link string) *AARConfigEntry {
	return &AARConfigEntry{
		Date:    date,
		Title:   title,
		Terrain: terrain,
		Link:    link,
	}
}

func ParseAARs(filedate string, aars []*AAR) []*AARConverted {
	// -- Start temp AAR parsing
	chans := make([]chan *AARConverted, 0, len(aars))
	for _, aar := range aars {
		if aar.exclude {
			continue
		}

		aar.date = filedate

		ch := make(chan *AARConverted)
		chans = append(chans, ch)

		go func() {
			defer close(ch)
			converted := aar.Parse()
			ch <- converted
		}()
	}

	// -- Gather converted AARs
	convertedAARs := make([]*AARConverted, 0)
	for _, ch := range chans {
		convertedAARs = append(convertedAARs, <-ch)
	}

	return convertedAARs
}

func UpdateAARListConfig(cfgPath string, entries []*AARConfigEntry) {
	// -- Read config
	file, err := os.Create("aarListConfig.tmp")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	cfg, err := os.Open(cfgPath)
	if err != nil {
		panic(err)
	}
	defer cfg.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString("aarConfig = [\n")

	for _, entry := range entries {
		out, err := json.MarshalIndent(entry, "    ", "    ")
		if err != nil {
			panic(err)
		}
		writer.WriteString("    ")
		writer.Write(out)
		writer.WriteString(",\n")
	}

	reader := bufio.NewScanner(cfg)
	reader.Scan()
	for reader.Scan() {
		writer.WriteString(reader.Text() + "\n")
	}
	writer.Flush()

	// -- Replace aarListConfig.ini with content of writter
	cfg.Close()
	os.Remove(cfg.Name())

	newCfg, err := os.Create(cfgPath)
	if err != nil {
		panic(err)
	}
	defer newCfg.Close()

	file.Seek(0, 0)
	if _, err := io.Copy(newCfg, file); err != nil {
		panic(err)
	}

	file.Close()
	os.Remove(file.Name())
	fmt.Println("Конфиг AAR обновлен.")
}

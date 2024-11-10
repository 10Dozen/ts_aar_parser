package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	AAR_TEST_META_PATTERN string = `<meta><core>`
	AAR_TEST_PATTERN      string = `<AAR-.*>`

	AAR_METADATA_PATTERN    string = `(.*) "<AAR-.*><meta><core>(.*)<\/core>`
	AAR_OBJECT_META_PATTERN string = `<meta><(unit|veh)>\{ ""(unit|veh)Meta"": (.*) \}<\/(unit|veh|av)>`
	AAR_FRAME_PATTERN       string = `<(\d+)><(unit|veh|av)>(.*)<\/(unit|veh|av)>`

	FLUSH_AFTER int = 10000
)

type AARRegexpRepo struct {
	test, testMeta,
	metadata, objectMetadata, frame *regexp.Regexp
}

type AARHandler struct {
	aars   []*AAR
	regexp *AARRegexpRepo
}

type AARConfigEntry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Terrain string `json:"terrain"`
	Link    string `json:"link"`
}

func (ah *AARHandler) ParseLine(line string) {
	// -- Check for AAR line
	if !ah.regexp.test.MatchString(line) {
		return
	}

	// -- Check for meta
	if ah.regexp.testMeta.MatchString(line) {
		matches := ah.regexp.metadata.FindStringSubmatch(line)
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
	// Force current buffer to flush
	aar.buff.Flush()
	aar.buff = nil
	aar.tmp.Close()
}

func (ah *AARHandler) ParseAARs(filedate string) {
	chans := make([]chan int, 0, len(ah.aars))

	for _, aar := range ah.aars {
		aar.date = filedate

		ch := make(chan int)
		chans = append(chans, ch)

		go func() {
			aar.Parse()
			ch <- 1
		}()
	}

	for _, ch := range chans {
		<-ch
	}
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

func (ah *AARHandler) UpdateConfig(cfgPath string, entries []*AARConfigEntry) {
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

func NewAARHandler() *AARHandler {
	h := &AARHandler{
		aars: make([]*AAR, 0),
		regexp: &AARRegexpRepo{
			test:           regexp.MustCompile(AAR_TEST_PATTERN),
			testMeta:       regexp.MustCompile(AAR_TEST_META_PATTERN),
			metadata:       regexp.MustCompile(AAR_METADATA_PATTERN),
			objectMetadata: regexp.MustCompile(AAR_OBJECT_META_PATTERN),
			frame:          regexp.MustCompile(AAR_FRAME_PATTERN),
		},
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

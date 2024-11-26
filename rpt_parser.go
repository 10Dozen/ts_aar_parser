package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	RPT_SUFFIX         string = ".rpt"
	RPT_VERSION_SUFFIX string = "x64"
)

type ReportContent struct {
	date   string   // date of parsed RPT file
	aars   []*AAR   // list of AAR caches extracted from RPT
	orbats []*ORBAT // orbat from RPT
}

func ParseLatestRPTs(path string) *ReportContent {
	date, reports := findLatestRPTs(path)
	if len(reports) == 0 {
		log.Fatalf("[ParseRPT] Failed to find any RPT file at %s", path)
	}

	// -- Process several .rpt file in parallel
	channels := make([]chan ReportContent, 0)
	for _, v := range reports {
		ch := make(chan ReportContent)
		channels = append(channels, ch)
		go parseRPT(v, path, ch)
	}

	// -- Gather data from goroutines into a single object
	content := &ReportContent{
		date:   date,
		aars:   make([]*AAR, 0),
		orbats: make([]*ORBAT, 0),
	}
	var fileContent ReportContent
	for _, ch := range channels {
		fileContent = <-ch

		content.date = fileContent.date
		content.aars = append(content.aars, fileContent.aars...)
		content.orbats = append(content.orbats, fileContent.orbats...)
	}

	return content
}

func parseRPT(filename string, path string, outChannel chan<- ReportContent) {
	content := ReportContent{}

	// -- Get date of the RPT
	parts := strings.Split(strings.ToLower(filename), "_")
	filedate := parts[2]
	if parts[1] != RPT_VERSION_SUFFIX {
		filedate = parts[1]
	}
	content.date = filedate

	// -- Read file
	file, err := os.Open(filepath.Join(path, filename))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// -- Add thread local handlers
	orbatHandler := NewORBATHandler()
	aarHandler := NewAARHandler()
	defer aarHandler.closeTmpReport()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		orbatHandler.ParseLine(line)
		aarHandler.ParseLine(line)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	aarHandler.closeTmpReport()

	content.aars = aarHandler.aars
	content.orbats = orbatHandler.orbats

	outChannel <- content
}

func findLatestRPTs(path string) (string, []string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	var (
		modTime time.Time
		date    string
		files   map[string][]string = make(map[string][]string)
	)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}

		// -- Skip not .rpt files
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".rpt") {
			continue
		}

		// -- Add files by date
		date = info.ModTime().Format("2006-01-02")
		filelist, ok := files[date]
		if !ok {
			filelist = make([]string, 0)
		}
		files[date] = append(filelist, info.Name())

		// -- Select latest file date
		if !(info.ModTime().Before(modTime)) {
			modTime = info.ModTime()
		}
	}

	// -- By latest file date get date to parse files from
	date = modTime.Format("2006-01-02")
	filesToParse := files[date]
	fmt.Printf("Свежайшие RPT файлы за %s (%d): \n", date, len(filesToParse))
	for _, v := range filesToParse {
		fmt.Printf("  - %s\n", v)
	}

	return date, filesToParse
}

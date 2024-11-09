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

func ParseRPT(path string, aarHandler *AARHandler, orbatHandler *ORBATHandler) (filedate string) {
	rpt_file := findLatestRPT(path)
	if rpt_file == "" {
		log.Fatalf(
			"[ParseRPT] Failed to find any RPT file at %s",
			path,
		)
	}

	parts := strings.Split(strings.ToLower(rpt_file), "_")
	filedate = parts[2]
	if parts[1] != RPT_VERSION_SUFFIX {
		filedate = parts[1]
	}

	file, err := os.Open(filepath.Join(path, rpt_file))
	if err != nil {
		panic(err)
	}
	defer file.Close()
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

	return
}

func findLatestRPT(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	var (
		modTime time.Time
		name    string
	)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}
		if !(info.ModTime().Before(modTime)) && strings.HasSuffix(strings.ToLower(entry.Name()), ".rpt") {
			modTime = info.ModTime()
			name = entry.Name()
		}
	}

	fmt.Printf("Свежайший RPT файл: %s\n", name)
	return name
}

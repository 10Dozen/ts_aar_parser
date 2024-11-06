package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const RPT_SUFFIX string = ".rpt"
const RPT_VERSION_SUFFIX string = "x64"

func ParseRPT(path string) (filedate string) {
	rpt_file := findLatestRPT(path)

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

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()
		// ParseORBATLine(line)
		ParseAARLine(line)
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

	fmt.Printf("RPT latest file: %s", name)
	return name
}

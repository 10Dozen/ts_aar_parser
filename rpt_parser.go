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

func ParseRPT(path string) {
	rpt_file := findLatestRPT(path)
	file, err := os.Open(rpt_file)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		line := scanner.Text()
		ParseORBATLine(line)
		// ParseAARLine(line)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
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
	return filepath.Join(path, name)
}

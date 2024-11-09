package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	AAR_TEST_META_PATTERN string = `<meta><core>`
	AAR_TEST_PATTERN      string = `<AAR-.*>`

	AAR_METADATA_PATTERN    string = `(.*) "<AAR-.*><meta><core>(.*)<\/core>`
	AAR_OBJECT_META_PATTERN string = `<AAR-.*><meta><(unit|veh)>\{ ""(unit|veh)Meta"": (.*) \}<\/(unit|veh|av)>`
	AAR_FRAME_PATTERN       string = `<AAR-.*><(\d+)><(unit|veh|av)>(.*)<\/(unit|veh|av)>`

	FLUSH_AFTER int = 10000
)

type AARRegexpRepo struct {
	test,
	metadata, objectMetadata, frame *regexp.Regexp
}

type AARHandler struct {
	aars   []*AAR
	regexp *AARRegexpRepo
}

func (ah *AARHandler) ParseLine(line string) {
	// -- Check for AAR line
	isAAR := ah.regexp.test.MatchString(line)
	if !isAAR {
		return
	}

	// -- Check for meta
	matches := ah.regexp.metadata.FindStringSubmatch(line)
	if matches != nil {
		//fmt.Println("AAR Metadata match")
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

	//fmt.Println("AAR data line match")
	ah.appendToTempReport(line)
}

func (ah *AARHandler) createTempReport(aar *AAR) {
	ah.closeTmpReport()
	/*TODO: Uncomment
	tmpFilepath := filepath.Join(
		configuration.ExecDirectory,
		fmt.Sprintf("%s.tmp", aar.Guid),
	)
	*/
	tmpFilepath := fmt.Sprintf("%s.tmp", aar.Guid)
	fmt.Printf("Creating temporary report %s\n", tmpFilepath)

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
	fmt.Println("Closing temporary report")
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
	for _, aar := range ah.aars {
		fmt.Printf("[AARHandler] Parsing AAR %s\n", aar.Guid)
		aar.date = filedate
		aar.Parse()
	}
}

func (ah *AARHandler) ToJSON(aar *AAR) string {
	outputData, err := json.MarshalIndent(aar.out, "", "    ") // TODO: json.Marshal(aar.out) //
	if err != nil {
		panic(err)
	}
	return string(outputData)
}

func (ah *AARHandler) Clear() {
	for _, aar := range ah.aars {
		ah.OmitAAR(aar)
	}
}

func (ah *AARHandler) OmitAAR(aar *AAR) {
	aar.tmp.Close()
	os.Remove(aar.tmp.Name())
	aar = nil
}

func NewAARHandler() *AARHandler {
	h := &AARHandler{
		aars: make([]*AAR, 0),
		regexp: &AARRegexpRepo{
			test:           regexp.MustCompile(AAR_TEST_PATTERN),
			metadata:       regexp.MustCompile(AAR_METADATA_PATTERN),
			objectMetadata: regexp.MustCompile(AAR_OBJECT_META_PATTERN),
			frame:          regexp.MustCompile(AAR_FRAME_PATTERN),
		},
	}

	return h
}

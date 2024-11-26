package main

import "regexp"

const (
	AAR_TEST_META_PATTERN   string = `<meta><core>`
	AAR_TEST_PATTERN        string = `<AAR-.*>`
	AAR_METADATA_PATTERN    string = `(.*) "<AAR-.*><meta><core>(.*)<\/core>`
	AAR_OBJECT_META_PATTERN string = `<meta><(unit|veh)>\{ ""(unit|veh)Meta"": (.*) \}<\/(unit|veh|av)>`
	AAR_FRAME_PATTERN       string = `<(\d+)><(unit|veh|av)>(.*)<\/(unit|veh|av)>`
	ORBAT_METADATA_PATTERN  string = `"\[tS_ORBAT\] Meta: (.*)"`
	ORBAT_DATA_PATTERN      string = `"\[tS_ORBAT\] (\[.*\])"`
)

type RegexpRepository struct {
	AAR   AARRegexRepository
	ORBAT ORBATRegexRepository
}

type AARRegexRepository struct {
	test, testMeta, metadata, objectMetadata, frame *regexp.Regexp
}

type ORBATRegexRepository struct {
	metadataRE, dataRE *regexp.Regexp
}

func NewRegexRepo() *RegexpRepository {
	repo := &RegexpRepository{
		AAR: AARRegexRepository{
			test:           regexp.MustCompile(AAR_TEST_PATTERN),
			testMeta:       regexp.MustCompile(AAR_TEST_META_PATTERN),
			metadata:       regexp.MustCompile(AAR_METADATA_PATTERN),
			objectMetadata: regexp.MustCompile(AAR_OBJECT_META_PATTERN),
			frame:          regexp.MustCompile(AAR_FRAME_PATTERN),
		},
		ORBAT: ORBATRegexRepository{
			metadataRE: regexp.MustCompile(ORBAT_METADATA_PATTERN),
			dataRE:     regexp.MustCompile(ORBAT_DATA_PATTERN),
		},
	}

	return repo
}

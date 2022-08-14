package db

import (
	"fmt"
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParse(t *testing.T) {
	file, err := ioutil.ReadFile("./example.yaml")
	if err != nil {
		t.Errorf("File reading example")
		return
	}
	parsed := make(map[string]MediaRecord)
	if err := yaml.Unmarshal(file, parsed); err != nil {
		t.Errorf("Failed unmarshalling: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf(fmt.Sprintf("Expected length of 2, received %d", len(parsed)))
	}

	ignored := parsed["252F2D6C07AE83BCD39C055288044B7624416A835C7F3856042051D80F35E5B8"]
	tagged := parsed["37EFA83C36884C46D5B09514B86AD6063555253060678E1773E3E13883313F1E"]
	ValidateIgnored(t, &ignored)
	ValidateTagged(t, &tagged)
}

func ValidateIgnored(t *testing.T, rec *MediaRecord) {
	if !*rec.Ignore {
		t.Errorf("Expected record to be ignored")
	}
	if !*rec.Reviewed {
		t.Errorf("Expected record to be reviewed")
	}
}

func ValidateTagged(t *testing.T, rec *MediaRecord) {

	date := rec.GetDate()
	if date.Year() != 2011 {
		t.Errorf("Expected year of 2011 for wedding picture")
	}

	if len(rec.Paths) != 2 {
		t.Errorf("Expected 2 paths to be loaded from YAML record")
	}

	if len(rec.Tags) != 1 {
		t.Errorf("Expected 1 tag to be loaded from YAML record")
	}
}

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

	date := parsed["37EFA83C36884C46D5B09514B86AD6063555253060678E1773E3E13883313F1E"].GetDate()
	if date.Year() != 2011 {
		t.Errorf("Expected year of 2011 for wedding picture")
	}
}

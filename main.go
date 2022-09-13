package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
	"gopkg.in/yaml.v3"
)

func main() {

	managedHelp := "Path to managed media directory"
	targetHelp := "Path to target reorg directory"
	var managed, target string
	flag.StringVar(&managed, "managed", ".", managedHelp)
	flag.StringVar(&managed, "m", ".", managedHelp+" (shorthand)")
	flag.StringVar(&target, "target", "", targetHelp)
	flag.StringVar(&target, "t", "", targetHelp+" (shorthand)")
	flag.Parse()
	records := loadRecordDatabase(managed)
	fmt.Printf("Record count: %d", len(records))
}

func loadRecordDatabase(path string) map[string]db.MediaRecord {
	file, err := ioutil.ReadFile(path + "/pic-man.db")
	parsed := make(map[string]db.MediaRecord)
	if err != nil {
		panic(fmt.Errorf("Parsing %s resulted in error %v", path, err))
	}
	if err := yaml.Unmarshal(file, parsed); err != nil {
		panic(fmt.Errorf("Parsing %s resulted in error %v", path, err))

	}
	return parsed
}

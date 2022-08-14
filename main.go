package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
	"gopkg.in/yaml.v3"
)

// **** Goals Below ****
// commands:
// - stats
// - reorg
// - add new folder
// - web
// jpeg metadata utility

var commands = [...]string{"stats", "reorg", "add", "web"}

func main() {

	managedHelp := "Path to managed media directory"
	comandHelp := commandOptionsToString() + ": Statistics, reorganize, add folder, and web API start."
	var managed, command string
	flag.StringVar(&managed, "managed", ".", managedHelp)
	flag.StringVar(&managed, "m", ".", managedHelp+" (shorthand)")
	flag.StringVar(&command, "command", "stats", comandHelp)
	flag.StringVar(&command, "c", "stats", comandHelp+" (shorthand)")
	flag.Parse()
	records := loadRecordDatabase(managed)
	fmt.Printf("Record count: %d", len(records))
}

func commandOptionsToString() string {
	result := ""
	for i, v := range commands {
		if i == 0 {
			result = v
		} else {
			result = fmt.Sprintf("%s|%s", result, v)
		}
	}
	return result
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

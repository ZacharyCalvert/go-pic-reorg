package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
	"github.com/ZacharyCalvert/go-pic-reorg/move"
	"gopkg.in/yaml.v3"
)

func main() {
	var dryRun bool
	var managed, target string
	parseFlags(&dryRun, &managed, &target)
	if dryRun {
		fmt.Printf("This is a dry run - no changes will be applied\n")
	}
	records := loadRecordDatabase(managed)
	validateDatabase(managed, records)
	validateTarget(target)
	mover := move.BuildMover(target, records)
	mover.PerformMove(managed, dryRun)
}

func validateTarget(target string) {
	_, err := os.Stat(target)
	if !(err != nil && errors.Is(err, os.ErrNotExist)) {
		fmt.Printf("Target directory %s must not exist", target)
		os.Exit(1)
	}
}

func validateDatabase(managed string, records map[string]db.MediaRecord) {
	// database validation is two steps:
	// 1: records non-empty (we should have loaded SOMETHING for a migration)
	// 2: each file must exist along the appropriate path

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Database validation failed: %v", err)
			os.Exit(1)
		}
	}()

	if len(records) == 0 {
		panic(errors.New("Database empty.  Aborting..."))
	}

	for id, rec := range records {
		if rec.IsIgnoredMedia() {
			continue
		}
		src := filepath.Join(managed, rec.StoredAt)
		if details, err := os.Stat(src); err != nil {
			panic(err)
		} else if details.IsDir() {
			panic(errors.New(fmt.Sprintf("Record %s is indicating a directory instead of a file at %s, indicating it is stored at %s", id, src, rec.StoredAt)))
		}
	}
}

func parseFlags(dryRun *bool, managed, target *string) {
	managedHelp := "Path to managed media directory"
	targetHelp := "Path to target reorg directory"
	dryHelp := "If this is a dry run (no side effects)"
	flag.BoolVar(dryRun, "dryrun", false, dryHelp)
	flag.BoolVar(dryRun, "d", false, dryHelp+" (shorthand)")
	flag.StringVar(managed, "managed", ".", managedHelp)
	flag.StringVar(managed, "m", ".", managedHelp+" (shorthand)")
	flag.StringVar(target, "target", "", targetHelp)
	flag.StringVar(target, "t", "", targetHelp+" (shorthand)")
	flag.Parse()
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

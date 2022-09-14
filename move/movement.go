package move

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
	"gopkg.in/yaml.v3"
)

// File movement requires no collision preparation, syncing on the act of movement (avoid races)
// and we want to support a dry run preparation
// we will track to->from targets
// resulting names and rebuild to avoid collision

type Mover interface {
	PerformMove(currentManaged string, dryRun bool)
}

type id string

type mover struct {
	records          map[string]db.MediaRecord
	idToDestination  map[id]string
	destinationsToID map[string]id
	target           string
}

// struggled a bit here given that outside interference could modify the existing
// db.MediaRecord.  Ideally, for isolation, this would be deep copied to enforce
// internal consistency.
// Generating a struct which exposes the "guts" from external force modification
// is an abhorrent break of encapsulation, but I'm making a concious choice here
// to deliver with speed as I'm the only expected customer.
// For folks looking into my pet project, this is not an ideal enterprise quality
// approach, and I acknowledge that openly.
func BuildMover(target string, records map[string]db.MediaRecord) Mover {
	idToDest := make(map[id]string) // where we will translate from a sha sum to a future home location for the media
	destToID := make(map[string]id) // means to track previous planned destinations
	for sha, rec := range records {
		if rec.IsIgnoredMedia() {
			continue
		}
		iteration := 0
		dest := generateDestFileName(rec.GetDate(), rec.Paths[0], iteration)

		// iterate until planned file path does not collide with existing planned destination
		for _, collides := destToID[dest]; collides; {
			iteration++
			dest = generateDestFileName(rec.GetDate(), rec.Paths[0], iteration)
			_, collides = destToID[dest]
		}
		idToDest[id(sha)] = dest
		destToID[dest] = id(sha)
	}
	return mover{records: records, idToDestination: idToDest, destinationsToID: destToID, target: target}
}

func (m mover) PerformMove(currentManaged string, dryRun bool) {
	for shaKey, record := range m.records {
		if record.IsIgnoredMedia() {
			continue
		}
		relativeFrom := record.StoredAt
		relativeDest, ok := m.idToDestination[id(shaKey)]
		from := filepath.Join(currentManaged, relativeFrom)
		dest := filepath.Join(m.target, relativeDest)
		if !ok {
			panic(fmt.Errorf("No destination planned for %s, currently stored at %s", shaKey, from))
		}
		if dryRun {
			fmt.Printf("Would move %s to %s for %d of time %v.\n", from, dest, record.Earliest, record.GetDate())
		} else {
			os.MkdirAll(filepath.Dir(dest), os.ModePerm)
			err := os.Rename(from, dest)
			if err != nil {
				panic(fmt.Errorf("Error moving %s to %s: %v", from, dest, err))
			}
		}
		record.StoredAt = dest
	}
	outDB := db.Database{LastUpdated: time.Now().Unix(), Media: m.records}
	outYML, err := yaml.Marshal(outDB)
	if err == nil {
		dbDest := filepath.Join(m.target, "pic-man.db")
		if !dryRun {
			os.WriteFile(dbDest, outYML, 0644)
		}
	} else {
		panic(fmt.Errorf("Could not prepare database yaml: %v", err))
	}
}

func generateDestFileName(date time.Time, previousPath string, iteration int) string {
	year, month, day := date.Date()
	importedFrom := forceSeparatorConsistency(previousPath)
	base := filepath.Base(importedFrom)
	parentDir := filepath.Base(filepath.Dir(importedFrom))
	relativeDir := fmt.Sprintf("%d/%02d/%02d", year, int(month), day)
	if parentDir != "" {
		relativeDir = fmt.Sprintf("%d/%02d/%02d/%s", year, int(month), day, parentDir)
	}
	if iteration == 0 {
		return fmt.Sprintf("%s/%s", relativeDir, base)
	}
	ext := filepath.Ext(previousPath)
	if len(ext) > 0 {
		return fmt.Sprintf("%s/%s_%d%s", relativeDir, base[:len(base)-len(ext)], iteration, ext)
	} else {
		return fmt.Sprintf("%s/%s_%d", relativeDir, filepath.Base(previousPath), iteration)
	}
}

func forceSeparatorConsistency(path string) string {
	path = strings.ReplaceAll(path, "/", fmt.Sprintf("%c", os.PathSeparator))
	path = strings.ReplaceAll(path, "\\", fmt.Sprintf("%c", os.PathSeparator))
	return path
}

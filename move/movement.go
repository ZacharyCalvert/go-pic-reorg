package move

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
	"gopkg.in/yaml.v3"
)

// File movement requires no collision preparation, syncing on the act of movement (avoid races)
// and we want to support a dry run preparation
// we will track to->from targets
// resulting names and rebuild to avoid collision

type Mover interface {
	PerformMove(dryRun bool)
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
		dest := generateDestFileName(rec.Earliest, rec.Paths[0], iteration)

		// iterate until planned file path does not collide with existing planned destination
		for _, collides := destToID[dest]; collides; {
			iteration++
			dest = generateDestFileName(rec.Earliest, rec.Paths[0], iteration)
			_, collides = destToID[dest]
		}
		idToDest[id(sha)] = dest
		destToID[dest] = id(sha)
	}
	return mover{records: records, idToDestination: idToDest, destinationsToID: destToID, target: target}
}

func (m mover) PerformMove(dryRun bool) {
	for shaKey, record := range m.records {
		if record.IsIgnoredMedia() {
			continue
		}
		from := record.StoredAt
		dest, ok := m.idToDestination[id(shaKey)]
		if !ok {
			panic(fmt.Errorf("No destination planned for %s, currently stored at %s", shaKey, from))
		}
		if dryRun {
			fmt.Printf("Would move %s to %s.\n", from, dest)
		} else {
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
		if dryRun {
			fmt.Printf("Database YML for %s:\n%s", dbDest, string(outYML))
		} else {
			os.WriteFile(dbDest, outYML, 644)
		}
	} else {
		panic(fmt.Errorf("Could not prepare database yaml: %v", err))
	}
}

func generateDestFileName(unixTime int64, previousPath string, iteration int) string {
	date := time.Unix(0, unixTime*int64(time.Nanosecond)).UTC()
	year, month, day := date.Date()
	base := filepath.Base(previousPath)
	if iteration == 0 {
		return fmt.Sprintf("%d/%02d/%02d/%s", year, int(month), day, base)
	}
	ext := filepath.Ext(previousPath)
	if len(ext) > 0 {
		return fmt.Sprintf("%d/%02d/%02d/%s_%d%s", year, int(month), day, base[:len(base)-len(ext)], iteration, ext)
	} else {
		return fmt.Sprintf("%d/%02d/%02d/%s_%d", year, int(month), day, filepath.Base(previousPath), iteration)
	}
}

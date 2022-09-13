package move

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ZacharyCalvert/go-pic-reorg/db"
)

// File movement requires no collision preparation, syncing on the act of movement (avoid races)
// and we want to support a dry run preparation
// we will track to->from targets
// resulting names and rebuild to avoid collision

type Mover interface {
}

type id string

type mover struct {
	records          map[string]db.MediaRecord
	idToDestination  map[id]string
	destinationsToID map[string]id
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
	idToDest := make(map[id]string)
	destToID := make(map[string]id)
	for sha, rec := range records {
		iteration := 0
		dest := generateDestFileName(rec.Earliest, rec.Paths[0], iteration)

		// iterate until planned file path does not collide
		for previousSha, exists := destToID[dest]; !exists; iteration++ {
			fmt.Printf("Collision detected at destination %s between sha identifier %s and %s", dest, previousSha, sha)
			dest = generateDestFileName(rec.Earliest, rec.Paths[0], iteration)
		}
		idToDest[id(sha)] = dest
		destToID[dest] = id(sha)
	}
	return mover{records: records, idToDestination: idToDest, destinationsToID: destToID}
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

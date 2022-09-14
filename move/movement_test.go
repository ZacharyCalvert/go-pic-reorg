package move

import (
	"fmt"
	"testing"
	"time"
)

var epochDay = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func TestFirstIteration(t *testing.T) {

	fmt.Printf("Unix: %v", epochDay)
	dest := generateDestFileName(epochDay, "/hello/world.jpg", 0)
	expected := "1970/01/01/hello/world.jpg"
	if expected != dest {
		t.Errorf("Expected %s but received %s", expected, dest)
	}
}
func TestPad(t *testing.T) {

	dest := generateDestFileName(epochDay, "/hello/world.jpg", 1)
	expected := "1970/01/01/hello/world_1.jpg"
	if expected != dest {
		t.Errorf("Expected %s but received %s", expected, dest)
	}
}
func TestPadNoExtension(t *testing.T) {

	dest := generateDestFileName(epochDay, "/hello/world", 1)
	expected := "1970/01/01/hello/world_1"
	if expected != dest {
		t.Errorf("Expected %s but received %s", expected, dest)
	}
}

package lookup

import (
	"testing"
)

func TestBibleAPIClient(t *testing.T) {
	RunLookupTests(t, NewBibleAPIClient())
}

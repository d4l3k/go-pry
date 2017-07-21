package pry

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func TestHistory(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	historyFile = ".go-pry_history_test"

	want := []string{
		"test",
		fmt.Sprintf("rand: %d", rand.Int63()),
	}
	saveHistory(&want)

	out := loadHistory()
	if !reflect.DeepEqual(want, out) {
		t.Errorf("loadHistory() = %+v; expected %+v", out, want)
	}
}
